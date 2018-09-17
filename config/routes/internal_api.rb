# frozen_string_literal: true

namespace :api do
  namespace :internal do
    resource :access_token, only: %i(create)
    resource :context_data, only: %i(show)
  end
end
