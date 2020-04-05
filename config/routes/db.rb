# frozen_string_literal: true

namespace :db do
  resource :search, only: [:show]

  resources :channel_groups, except: %i(show) do
    member do
      patch :unpublish
    end
  end

  resources :channels, except: %i(show) do
    member do
      patch :hide
    end
  end

  resources :characters, except: [:show] do
    member do
      get :activities
      patch :hide
    end
  end

  resources :casts, only: %i(edit update destroy) do
    member do
      get :activities
      patch :hide
    end
  end

  resources :comments, only: %i(create destroy)

  resources :episodes, only: %i(edit update destroy) do
    member do
      get :activities
      patch :hide
    end
  end

  resources :organizations, except: [:show] do
    member do
      get :activities
      patch :hide
    end
  end

  resources :people, except: [:show] do
    member do
      get :activities
      patch :hide
    end
  end

  resources :slots, only: %i(edit update destroy) do
    member do
      get :activities
      patch :hide
    end
  end

  resources :trailers, only: %i(edit update destroy) do
    member do
      get :activities
      patch :hide
    end
  end

  resources :staffs, only: %i(edit update destroy) do
    member do
      get :activities
      patch :hide
    end
  end

  resources :programs, only: %i(edit update destroy) do
    member do
      get :activities
      patch :hide
    end
  end

  resources :vod_titles, only: %i(index) do
    member do
      patch :hide
    end
  end

  resources :works, except: [:show] do
    collection do
      get :season
      get :resourceless
    end

    member do
      get :activities
      patch :hide
    end

    resource :image, controller: :work_images, only: %i(show create update destroy)
    resources :casts, only: %i(index new create)
    resources :episodes, only: %i(index new create)
    resources :slots, only: %i(index new create)
    resources :trailers, only: %i(index new create)
    resources :staffs, only: %i(index new create)
    resources :programs, only: %i(index new create)
  end

  root "home#index"
end

scope module: :db do
  constraints format: "html" do
    # rubocop:disable Layout/ExtraSpacing, Layout/LineLength
    match "/db/activities",                         via: :get,    as: :db_activity_list,          to: "activities#index"
    match "/db/series",                             via: :get,    as: :db_series_list,            to: "series#index"
    match "/db/series",                             via: :post,                                   to: "series#create"
    match "/db/series/:id",                         via: :delete, as: :db_series_detail,          to: "series#destroy"
    match "/db/series/:id",                         via: :patch,                                  to: "series#update"
    match "/db/series/:id/edit",                    via: :get,    as: :db_edit_series,            to: "series#edit"
    match "/db/series/:id/publishing",              via: :delete, as: :db_series_publishing,      to: "series_publishings#destroy"
    match "/db/series/:id/publishing",              via: :post,                                   to: "series_publishings#create"
    match "/db/series/:series_id/series_works",     via: :get,    as: :db_series_work_list,       to: "series_works#index"
    match "/db/series_works/:id",                   via: :delete, as: :db_series_work_detail,     to: "series_works#destroy"
    match "/db/series_works/:id",                   via: :patch,                                  to: "series_works#update"
    match "/db/series_works/:id/edit",              via: :get,    as: :db_edit_series_work,       to: "series_works#edit"
    match "/db/series_works/:id/publishing",        via: :delete, as: :db_series_work_publishing, to: "series_work_publishings#destroy"
    match "/db/series_works/:id/publishing",        via: :post,                                   to: "series_work_publishings#create"
    match "/db/series/:series_id/series_works/new", via: :get,    as: :db_new_series_work,        to: "series_works#new"
    match "/db/series/new",                         via: :get,    as: :db_new_series,             to: "series#new"
    # rubocop:enable Layout/ExtraSpacing, Layout/LineLength
  end
end
