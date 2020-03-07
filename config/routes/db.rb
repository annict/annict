# frozen_string_literal: true

scope module: :db do
  constraints(format: "html") do
    delete "/db/works/:id/appearance", to: "work_appearances#destroy", as: :db_works_appearance
    post   "/db/works/:id/appearance", to: "work_appearances#create"
    get    "/db/works/:id/edit", to: "works#edit", as: :db_works_edit
    delete "/db/works/:id", to: "works#destroy", as: :db_works_detail
    get    "/db/works/new", to: "works#new", as: :db_works_new
    get    "/db/works", to: "works#index", as: :db_works
    get    "/db", to: "home#show", as: :db
  end
end
