# frozen_string_literal: true

namespace :api do
  namespace :internal do
    post :graphql, to: "graphql#execute"

    resource :impression, only: %i(show update)
    resource :privacy_policy_agreement, only: %i(create)
    resource :records_sort_type, only: %i(update)
    resource :search, only: [:show]
    resources :characters, only: [:index]
    resources :mute_users, only: [:create]
    resources :organizations, only: [:index]
    resources :people, only: [:index]
    resources :receptions, only: %i(create destroy)
    resources :series_list, only: %i(index)
    resources :works, only: %i(index show)

    resources :favorites, only: %i(create) do
      post :unfavorite, on: :collection
    end

    resources :follows, only: %i(create) do
      post :unfollow, on: :collection
    end

    resources :likes, only: [:create] do
      post :unlike, on: :collection
    end

    resources :reactions, only: [] do
      post :add, on: :collection
      post :remove, on: :collection
    end

    resources :statistics, only: [] do
      get :user_heatmap, on: :collection
    end

    resource :user, only: [] do
      resources :slots, only: [:index], controller: "user_slots"
    end

    resources :works, only: [] do
      resource :library_entry, only: [:show]
      resource :status_chart_data, only: %i(show)
      resource :watchers_chart_data, only: %i(show)

      resources :statuses, only: [] do
        post :select, on: :collection
      end
    end
  end
end

scope module: :api do
  scope module: :internal do
    constraints format: "json" do
      # rubocop:disable Layout/ExtraSpacing, Layout/LineLength
      match "/api/internal/episode_records",                   via: :post,   as: :internal_api_episode_record_list,     to: "episode_records#create"
      match "/api/internal/following",                         via: :get,    as: :internal_api_following_list,          to: "following#index"
      match "/api/internal/library_entries",                   via: :get,    as: :internal_api_library_entry_list,      to: "library_entries#index"
      match "/api/internal/library_entries/:library_entry_id", via: :patch,  as: :internal_api_library_entry,           to: "library_entries#update"
      match "/api/internal/likes",                             via: :get,    as: :internal_api_like_list,               to: "likes#index"
      match "/api/internal/multiple_episode_records",          via: :post,   as: :internal_api_multiple_episode_record, to: "multiple_episode_records#create"
      match "/api/internal/skipped_episodes",                  via: :post,   as: :internal_api_skipped_episode_list,    to: "skipped_episodes#create"
      match "/api/internal/tracked_resources",                 via: :get,    as: :internal_api_tracked_resource_list,   to: "tracked_resources#index"
      match "/api/internal/user",                              via: :get,    as: :internal_api_user_detail,             to: "users#show"
      match "/api/internal/work_friends",                      via: :get,    as: :internal_api_work_friend_list,        to: "work_friends#index"
      # rubocop:enable Layout/ExtraSpacing, Layout/LineLength
    end

    constraints format: "html" do
      # rubocop:disable Layout/ExtraSpacing, Layout/LineLength
      match "/api/internal/activities",                 via: :get, as: :internal_api_activity_list,              to: "activities#index"
      match "/api/internal/bulk_operations/:job_id",    via: :get, as: :internal_api_bulk_operation_list,        to: "bulk_operations#show"
      match "/api/internal/program_list_modal_content", via: :get, as: :internal_api_program_list_modal_content, to: "program_list_modal_contents#show"
      # rubocop:enable Layout/ExtraSpacing, Layout/LineLength
    end
  end
end
