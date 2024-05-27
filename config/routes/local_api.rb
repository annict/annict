# typed: false
# frozen_string_literal: true

if Rails.env.development?
  scope module: :v4 do
    scope module: :local_api do
      constraints(format: "json") do
        post "/api/local/graphql", to: "graphql#execute"
      end
    end
  end
end
