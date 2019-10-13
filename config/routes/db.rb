# frozen_string_literal: true

namespace :db do
  resources :activities, only: [:index]
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

  resources :programs, only: %i(edit update destroy) do
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

  resources :series, only: %i(index new create edit update destroy) do
    member do
      get :activities
      patch :hide
    end

    resources :series_works, only: %i(index new create)
  end

  resources :series_works, only: %i(edit update destroy) do
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

  resources :program_details, only: %i(edit update destroy) do
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
    resources :programs, only: %i(index new create)
    resources :trailers, only: %i(index new create)
    resources :staffs, only: %i(index new create)
    resources :program_details, only: %i(index new create)
  end

  root "home#index"
end
