-- +goose Up
CREATE TABLE contact_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,

    email TEXT NOT NULL,

    contact_department_id UUID NOT NULL,

    message VARCHAR(500) NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT message_length_check CHECK (char_length(message) <= 500),
    CONSTRAINT fk_contact_messages_contact_departments
      FOREIGN KEY (contact_department_id)
      REFERENCES contact_departments(id)
      ON DELETE CASCADE

   
);

CREATE INDEX idx_contact_messages_email_created_at
ON contact_messages (email, created_at DESC);

-- +goose Down
DROP TABLE contact_messages;