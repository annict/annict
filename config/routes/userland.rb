# frozen_string_literal: true

namespace :userland do
  resources :projects, except: %i[index]

  root "home#index"
end
