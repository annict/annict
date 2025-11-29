# typed: false
# frozen_string_literal: true

scope module: :api do
  constraints(subdomain: "api") do
    post :graphql, to: "graphql#execute"

    namespace :canary do
      post :graphql, to: "graphql#execute"
    end

    namespace :v1 do
      resources :activities, only: %i[index]
      resources :casts, only: %i[index]
      resources :characters, only: %i[index]
      resources :episodes, only: [:index]
      resources :followers, only: %i[index]
      resources :following, only: %i[index]
      resources :organizations, only: %i[index]
      resources :people, only: %i[index]
      resources :records, only: [:index]
      resources :reviews, only: %i[index]
      resources :series, only: %i[index]
      resources :staffs, only: %i[index]
      resources :users, only: %i[index]
      resources :works, only: [:index]

      namespace :me do
        resources :following_activities, only: %i[index]
        resources :programs, only: %i[index], controller: :slots
        resources :records, only: %i[create update destroy]
        resources :reviews, only: %i[create update destroy]
        resources :statuses, only: [:create]
        resources :works, only: [:index]

        root "index#show"
      end

      match "*path", to: "application#not_found", via: :all
    end
  end
end
