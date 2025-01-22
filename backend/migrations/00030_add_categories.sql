-- +goose Up
-- +goose StatementBegin
CREATE TABLE public.categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
)
CREATE TABLE pubic.open_content_types (
    category_id integer NOT NULL,
    content_id integer NOT NULL,
    open_content_provider_id integer NOT NULL,
    PRIMARY KEY (category_id, content_id, open_content_provider_id),
    FOREIGN KEY (category_id) REFERENCES public.categories(id),
    FOREIGN KEY (content_id) REFERENCES public.libraries(id),
    FOREIGN KEY (content_id) REFERENCES public.videos(id),
    FOREIGN KEY (content_id) REFERENCES public.helpful_links(id),
    FOREIGN KEY (open_content_provider_id) REFERENCES open_content_providers(id);
)
INSERT INTO categories (name) VALUES ('Adult Basic Education')
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
