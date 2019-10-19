# frozen_string_literal: true

namespace :api do
  namespace :internal do
    post :graphql, to: "graphql#execute"

    resource :impression, only: %i(show update)
    resource :privacy_policy_agreement, only: %i(create)
    resource :slots_sort_type, only: [:update]
    resource :records_sort_type, only: %i(update)
    resource :search, only: [:show]
    resources :activities, only: [:index]
    resources :app_data, only: %i(index)
    resources :characters, only: [:index]
    resources :mute_users, only: [:create]
    resources :organizations, only: [:index]
    resources :page_data, only: %i(index)
    resources :people, only: [:index]
    resources :receptions, only: %i(create destroy)
    resources :series_list, only: %i(index)
    resources :works, only: %i(index show)

    resources :episodes, only: [] do
      resources :records, only: %i(create), controller: :episode_records
    end

    resources :favorites, only: %i(create) do
      post :unfavorite, on: :collection
    end

    resources :follows, only: %i(create) do
      post :unfollow, on: :collection
    end

    resources :latest_statuses, only: [] do
      patch :skip_episode
    end

    resources :likes, only: [:create] do
      post :unlike, on: :collection
    end

    resources :multiple_records, only: %i(create)

    resources :tips, only: [] do
      post :close, on: :collection
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
      resource :latest_status, only: [:show]
      resource :status_chart_data, only: %i(show)
      resource :watchers_chart_data, only: %i(show)

      resources :channels, only: [] do
        post :select, on: :collection
      end

      resources :statuses, only: [] do
        post :select, on: :collection
      end
    end
  end
end
