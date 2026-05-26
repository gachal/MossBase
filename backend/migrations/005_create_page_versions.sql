CREATE TABLE IF NOT EXISTS page_versions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    page_id BIGINT UNSIGNED NOT NULL,
    version_number INT NOT NULL,
    title VARCHAR(255) DEFAULT '',
    content LONGTEXT,
    content_html LONGTEXT,
    edited_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_page_id (page_id),
    INDEX idx_page_version (page_id, version_number),
    CONSTRAINT fk_version_page FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE,
    CONSTRAINT fk_version_editor FOREIGN KEY (edited_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
