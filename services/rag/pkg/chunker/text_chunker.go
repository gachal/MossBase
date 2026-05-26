package chunker

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gachal/mossbase/services/rag/internal/domain/entity"
	"github.com/gachal/mossbase/services/rag/internal/infrastructure/config"
	"github.com/google/uuid"
)

const (
	defaultChunkSize    = 512
	defaultChunkOverlap = 64
)

// codeBlockRe matches fenced code blocks (```...```) including the fences.
var codeBlockRe = regexp.MustCompile("(?s)```.*?```")

// TextChunker splits text into overlapping chunks, preserving code blocks as
// atomic units and using paragraph boundaries as split points.
type TextChunker struct {
	ChunkSize    int
	ChunkOverlap int
}

// NewTextChunker creates a TextChunker from the given configuration.
// Falls back to sensible defaults when config values are zero.
func NewTextChunker(cfg config.ChunkerConfig) *TextChunker {
	chunkSize := cfg.ChunkSize
	if chunkSize <= 0 {
		chunkSize = defaultChunkSize
	}

	chunkOverlap := cfg.ChunkOverlap
	if chunkOverlap <= 0 {
		chunkOverlap = defaultChunkOverlap
	}

	return &TextChunker{
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
	}
}

// Chunk splits the given content into overlapping chunks. Each chunk's Content
// field is prefixed with the title for contextual relevance.
func (c *TextChunker) Chunk(title string, content string) []entity.Chunk {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil
	}

	segments := c.extractSegments(trimmed)
	if len(segments) == 0 {
		return nil
	}

	return c.buildChunks(title, segments)
}

// segment represents a portion of text that should be kept together when
// building chunks. Code blocks are always atomic; prose may be split.
type segment struct {
	text    string
	isCode  bool
	charLen int
}

// extractSegments replaces code blocks with placeholders, splits the remaining
// prose by paragraph boundaries, then re-inserts the code blocks as atomic
// segments in their original order.
func (c *TextChunker) extractSegments(content string) []segment {
	// Collect all code blocks and replace with unique placeholders.
	var codeBlocks []string
	placeholder := func(match string) string {
		idx := len(codeBlocks)
		codeBlocks = append(codeBlocks, match)
		return fmt.Sprintf("\x00CODEBLOCK%d\x00", idx)
	}

	processed := codeBlockRe.ReplaceAllStringFunc(content, placeholder)

	// Split processed text by paragraph boundaries.
	parts := strings.Split(processed, "\n\n")

	var segments []segment
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		// Check if this part contains a code-block placeholder.
		if strings.Contains(trimmed, "\x00") {
			// Expand placeholders back into atomic code-block segments,
			// interleaving any surrounding prose.
			segments = expandPlaceholders(trimmed, codeBlocks, segments)
		} else {
			segments = append(segments, segment{
				text:    trimmed,
				isCode:  false,
				charLen: len(trimmed),
			})
		}
	}

	return segments
}

// expandPlaceholders takes a string that may contain one or more code-block
// placeholders, splits around them, and appends the resulting segments.
func expandPlaceholders(part string, codeBlocks []string, segments []segment) []segment {
	var buf strings.Builder
	i := 0
	for i < len(part) {
		if part[i] == '\x00' {
			// Flush any buffered prose before the placeholder.
			if buf.Len() > 0 {
				prose := strings.TrimSpace(buf.String())
				if prose != "" {
					segments = append(segments, segment{
						text:    prose,
						isCode:  false,
						charLen: len(prose),
					})
				}
				buf.Reset()
			}
			// Find the matching closing \x00 and extract the index.
			j := i + 1
			for j < len(part) && part[j] != '\x00' {
				j++
			}
			// Parse index between markers.
			idxStr := part[i+1 : j]
			var idx int
			fmt.Sscanf(idxStr, "CODEBLOCK%d", &idx)
			if idx < len(codeBlocks) {
				code := codeBlocks[idx]
				segments = append(segments, segment{
					text:    code,
					isCode:  true,
					charLen: len(code),
				})
			}
			i = j + 1
		} else {
			buf.WriteByte(part[i])
			i++
		}
	}

	// Flush trailing prose.
	if buf.Len() > 0 {
		prose := strings.TrimSpace(buf.String())
		if prose != "" {
			segments = append(segments, segment{
				text:    prose,
				isCode:  false,
				charLen: len(prose),
			})
		}
	}

	return segments
}

// buildChunks merges segments into chunks respecting ChunkSize, then applies
// overlap from the tail of the previous chunk.
func (c *TextChunker) buildChunks(title string, segments []segment) []entity.Chunk {
	var chunks []entity.Chunk
	var currentParts []string
	currentLen := 0
	chunkIndex := 0

	flush := func() {
		if len(currentParts) == 0 {
			return
		}
		body := strings.Join(currentParts, "\n\n")
		fullContent := title + "\n\n" + body

		if strings.TrimSpace(fullContent) == "" {
			return
		}

		now := time.Now()
		chunks = append(chunks, entity.Chunk{
			ID:         uuid.New().String(),
			ChunkIndex: chunkIndex,
			Title:      title,
			Content:    fullContent,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
		chunkIndex++

		// Compute overlap text from the body of this chunk.
		overlapText := c.overlapSuffix(body)

		// Reset for next chunk, starting with overlap.
		currentParts = nil
		currentLen = 0
		if overlapText != "" {
			currentParts = append(currentParts, overlapText)
			currentLen = len(overlapText)
		}
	}

	for _, seg := range segments {
		// If the segment alone exceeds ChunkSize, flush current buffer first.
		if seg.charLen > c.ChunkSize && currentLen > 0 {
			flush()
		}

		// For code blocks or oversized segments that cannot be split,
		// add them as-is even if they exceed ChunkSize.
		if seg.isCode || seg.charLen > c.ChunkSize {
			if currentLen > 0 {
				flush()
			}
			currentParts = append(currentParts, seg.text)
			currentLen += seg.charLen
			flush()
			continue
		}

		// Check if adding this segment would exceed the limit.
		additionalLen := seg.charLen
		if len(currentParts) > 0 {
			additionalLen += 2 // for the "\n\n" separator
		}

		if currentLen+additionalLen > c.ChunkSize {
			flush()
		}

		currentParts = append(currentParts, seg.text)
		currentLen += additionalLen
	}

	// Flush remaining buffer.
	if len(currentParts) > 0 {
		remaining := strings.Join(currentParts, "\n\n")
		fullContent := title + "\n\n" + remaining
		if strings.TrimSpace(fullContent) != "" {
			now := time.Now()
			chunks = append(chunks, entity.Chunk{
				ID:         uuid.New().String(),
				ChunkIndex: chunkIndex,
				Title:      title,
				Content:    fullContent,
				CreatedAt:  now,
				UpdatedAt:  now,
			})
		}
	}

	return chunks
}

// overlapSuffix returns the last ~ChunkOverlap characters from the body,
// trimmed to the nearest paragraph boundary so we don't split mid-sentence.
func (c *TextChunker) overlapSuffix(body string) string {
	if c.ChunkOverlap <= 0 || len(body) <= c.ChunkOverlap {
		return ""
	}

	tail := body[len(body)-c.ChunkOverlap:]

	// Try to snap forward to the first paragraph break.
	if idx := strings.Index(tail, "\n\n"); idx >= 0 {
		tail = tail[idx+2:] // skip past the double newline
	}

	tail = strings.TrimSpace(tail)
	return tail
}
