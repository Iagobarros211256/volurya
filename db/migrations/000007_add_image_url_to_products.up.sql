ALTER TABLE products ADD COLUMN IF NOT EXISTS image_url TEXT;




/*


image_url TEXT sem validação de formato
Qualquer string é aceita — incluindo valores inválidos. A validação de URL fica inteiramente na aplicação. Não é possível fazer validação de URL no PostgreSQL sem extensão, mas vale documentar a convenção esperada:
sqlALTER TABLE products ADD COLUMN IF NOT EXISTS image_url TEXT
    CHECK (image_url IS NULL OR image_url LIKE 'https://%');

🟡 Sem índice se houver busca por produtos com/sem imagem
sqlCREATE INDEX idx_products_image_url ON products (image_url) WHERE image_url IS NULL;
Útil para jobs que processam produtos sem imagem.

🟢 Coluna nullable por padrão — intencional mas vale documentar
Produtos existentes ficam com image_url = NULL, o que é correto para uma migration additive. Garanta que a aplicação trata NULL como "sem imagem" de forma consistente.

*/