# AGENTS.md file

## Using minimum requirements
- Use go `1.25.3`
- Use node `22`
- Use python `3.9.22`
- Use elixir `1.19.0`
- Use formatter: go fmt, prettier, mix format or black for python

## PR & Branch Instructions
- Valid branch naming feat/, fix/, chore/
    - example `feat/add-js-sample`, `fix/remove-request`, `chore/update-docs`
- PR title: `<type>(<scope>): <description>`
    - example `chore(docs): add title`

## Task
- Create a binance collector data and store it to file, the API should be 
`https://www.binance.com/bapi/asset/v2/public/asset-service/product/get-product-by-symbol?symbol=${symbol}`
- Append it to ndjson with hourly rotation
- Create a folder go, node, python and elixir
- Create a unittest for each language