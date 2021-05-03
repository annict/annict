# frozen_string_literal: true

namespace :forum do
  resources :categories, only: %i[show]

  resources :posts, only: %i[new create show edit update] do
    resources :comments, only: %i[new create edit update]
  end

  root "home#index"
end
