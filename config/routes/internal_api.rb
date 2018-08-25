# frozen_string_literal: true

namespace :api do
  namespace :internal do
    namespace :v3 do
      resource :access_token, only: %i(create)
      resource :base_data, only: %i(show)
    end

    resource :impression, only: %i(show update)
    resource :programs_sort_type, only: %i(update)
    resource :records_sort_type, only: %i(update)
    resource :search, only: %i(show)

    resources :activities, only: [:index]
    resources :app_data, only: %i(index)
    resources :characters, only: [:index]
    resources :items, only: %i(create)
    resources :mute_users, only: [:create]
    resources :organizations, only: [:index]
    resources :page_data, only: %i(index)
    resources :people, only: [:index]
    resources :receptions, only: %i(create destroy)
    resources :series_list, only: %i(index)
    resources :works, only: %i(index show)

    resource :amazon, only: [], controller: :amazon do
      get :search
    end

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
      resources :programs, only: [:index], controller: "user_programs"
    end

    resources :works, only: [] do
      resource :latest_status, only: [:show]

      resources :channels, only: [] do
        post :select, on: :collection
      end

      resources :statuses, only: [] do
        post :select, on: :collection
      end
    end
  end
end
