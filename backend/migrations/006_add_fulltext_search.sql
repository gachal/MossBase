-- 006: Add FULLTEXT index for page search (supports Chinese via ngram parser)
ALTER TABLE pages ADD FULLTEXT INDEX ft_page_search (title, content) WITH PARSER ngram;
