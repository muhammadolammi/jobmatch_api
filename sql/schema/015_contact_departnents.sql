
-- +goose Up
CREATE TABLE contact_departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT  NOT NULL
   
);

INSERT INTO contact_departments (
name  )
VALUES ( 'General Inquiry' );

INSERT INTO contact_departments (
name  )
VALUES ( 'Sales' );

INSERT INTO contact_departments (
name  )
VALUES ( 'Technical Support' );

INSERT INTO contact_departments (
name  )
VALUES ( 'Partnerships' );

-- +goose Down
DROP TABLE contact_departments;