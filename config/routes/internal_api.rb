# frozen_string_literal: true

namespace :api do
  namespace :internal do
    resources :mute_users, only: [:create]

    resources :reactions, only: [] do
      post :add, on: :collection
      post :remove, on: :collection
    end

    resource :user, only: [] do
      resources :slots, only: [:index], controller: "user_slots"
    end

    resources :works, only: [] do
      resource :library_entry, only: [:show]
    end
  end
end

scope module: :api do
  scope module: :internal do
    constraints format: "json" do
      # rubocop:disable Layout/ExtraSpacing, Layout/LineLength
      match "/api/internal/episode_records",                   via: :post,   as: :internal_api_episode_record_list,     to: "episode_records#create"
      match "/api/internal/multiple_episode_records",          via: :post,   as: :internal_api_multiple_episode_record, to: "multiple_episode_records#create"
      match "/api/internal/received_channels",                 via: :get,    as: :internal_api_received_channel_list,   to: "received_channels#index"
      # rubocop:enable Layout/ExtraSpacing, Layout/LineLength
    end
  end
end
