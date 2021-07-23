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
      match "/api/internal/library_entries/:library_entry_id", via: :patch,  as: :internal_api_library_entry,           to: "library_entries#update"
      match "/api/internal/multiple_episode_records",          via: :post,   as: :internal_api_multiple_episode_record, to: "multiple_episode_records#create"
      match "/api/internal/received_channels",                 via: :get,    as: :internal_api_received_channel_list,   to: "received_channels#index"
      match "/api/internal/registrations",                     via: :post,   as: :internal_api_registrations,           to: "registrations#create"
      match "/api/internal/sign_in",                           via: :post,   as: :internal_api_sign_in,                 to: "sign_in#create"
      match "/api/internal/sign_up",                           via: :post,   as: :internal_api_sign_up,                 to: "sign_up#create"
      match "/api/internal/skipped_episodes",                  via: :post,   as: :internal_api_skipped_episode_list,    to: "skipped_episodes#create"
      match "/api/internal/tracked_resources",                 via: :get,    as: :internal_api_tracked_resource_list,   to: "tracked_resources#index"
      match "/api/internal/user",                              via: :get,    as: :internal_api_user_detail,             to: "users#show"
      match "/api/internal/work_friends",                      via: :get,    as: :internal_api_work_friend_list,        to: "work_friends#index"
      # rubocop:enable Layout/ExtraSpacing, Layout/LineLength
    end
  end
end
