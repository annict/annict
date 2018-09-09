# frozen_string_literal: true

namespace :api do
  namespace :internal do
    resource :access_token, only: %i(create)
    resource :base_data, only: %i(show)
  end
end
