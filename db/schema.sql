CREATE DATABASE contentforge;
GO

USE contentforge;
GO

CREATE TABLE users (
    id INT IDENTITY(1,1) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT GETDATE()
);

CREATE TABLE media_files (
    id INT IDENTITY(1,1) PRIMARY KEY,
    user_id INT FOREIGN KEY REFERENCES users(id),
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Uploaded', 
    created_at DATETIME DEFAULT GETDATE(),
    CONSTRAINT CHK_MediaStatus CHECK (status IN ('Uploaded', 'Transcribing', 'Transcribed', 'Generating', 'Completed', 'Failed'))
);

CREATE TABLE generation_requests (
    id INT IDENTITY(1,1) PRIMARY KEY,
    user_id INT FOREIGN KEY REFERENCES users(id),
    media_file_id INT FOREIGN KEY REFERENCES media_files(id),
    prompt_modifier VARCHAR(max),
    created_at DATETIME DEFAULT GETDATE()
);

CREATE TABLE transcripts (
    id INT IDENTITY(1,1) PRIMARY KEY,
    media_file_id INT UNIQUE FOREIGN KEY REFERENCES media_files(id),
    raw_text NVARCHAR(max) NOT NULL,
    created_at DATETIME DEFAULT GETDATE()
);

CREATE TABLE generated_content (
    id INT IDENTITY(1,1) PRIMARY KEY,
    request_id INT UNIQUE FOREIGN KEY REFERENCES generation_requests(id),
    media_file_id INT FOREIGN KEY REFERENCES media_files(id),
    content_type VARCHAR(50),
    result_text NVARCHAR(max) NOT NULL,
    created_at DATETIME DEFAULT GETDATE()
);
GO