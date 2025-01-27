-- +goose Up
-- +goose StatementBegin
CREATE TABLE public.open_content_categories (
    key SERIAL PRIMARY KEY,
    value VARCHAR(255) NOT NULL
);
CREATE INDEX idx_open_content_categories_value ON public.open_content_categories USING btree(value);

CREATE TABLE public.open_content_types (
    category_key integer NOT NULL,
    content_id integer NOT NULL,
    open_content_provider_id integer NOT NULL,
    PRIMARY KEY (category_key, content_id, open_content_provider_id) 
);

INSERT INTO open_content_categories (name) VALUES 
('Adult Basic Education'),
('Higher Education'),
('Vocational'),
('Rehabilitative'),
('Life Skills'),
('Therapeutic');

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.open_content_categories CASCADE; 
DROP TABLE IF EXISTS public.open_content_types CASCADE;
-- +goose StatementEnd


