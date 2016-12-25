# frozen_string_literal: true

Rails.application.routes.draw do
  if Rails.env.development?
    mount LetterOpenerWeb::Engine, at: "/low"
    mount Dmmyix::Engine, at: "/dmmyix"
  end

  get "dummy_image", to: "application#dummy_image" if Rails.env.test?

  devise_for :users,
    controllers: { omniauth_callbacks: :callbacks },
    skip: %i(passwords registrations sessions)

  devise_scope :user do
    get "sign_up", to: "registrations#new", as: :new_user_registration
    get "sign_in", to: "sessions#new", as: :new_user_session
    post "sign_in", to: "sessions#create", as: :user_session
    post "users", to: "registrations#create", as: :user_registration
    delete "sign_out", to: "sessions#destroy", as: :destroy_user_session
    resource :password, only: %i(new create edit update)
    resources :oauth_users, only: %i(new create)
  end

  use_doorkeeper do
    controllers applications: "oauth/applications",
                authorizations: "oauth/authorizations"
    skip_controllers :authorized_applications
  end

  namespace :api do
    namespace :internal do
      resource :programs_sort_type, only: [:update]
      resource :search, only: [:show]
      resources :activities, only: [:index]
      resources :characters, only: [:index]
      resources :mute_users, only: [:create]
      resources :organizations, only: [:index]
      resources :people, only: [:index]
      resources :receptions, only: [:create, :destroy]

      resources :dislikes, only: [:create] do
        post :undislike, on: :collection
      end

      resources :follows, only: %i(create) do
        post :unfollow, on: :collection
      end

      resources :latest_statuses, only: [:index] do
        patch :skip_episode
      end

      resources :likes, only: [:create] do
        post :unlike, on: :collection
      end

      resources :multiple_records, only: %i(create)

      resources :tips, only: [] do
        post :close, on: :collection
      end

      resources :records, only: %i(create) do
        get :user_heatmap, on: :collection
      end

      resource :user, only: [] do
        resources :programs, only: [:index], controller: "user_programs"
      end

      resources :works, only: [] do
        get :friends

        resource :latest_status, only: [:show]
        resources :images, controller: :work_images, only: %i(create)

        resources :channels, only: [] do
          post :select, on: :collection
        end
      end
    end
  end

  scope module: :api do
    constraints(subdomain: "api") do
      namespace :v1 do
        resources :episodes, only: [:index]
        resources :records, only: [:index]
        resources :works, only: [:index]

        namespace :me do
          resources :programs, only: [:index]
          resources :records, only: [:create, :update, :destroy]
          resources :statuses, only: [:create]
          resources :works, only: [:index]
        end

        match "*path", to: "application#not_found", via: :all
      end
    end
  end

  scope module: :db, as: :db do
    constraints(subdomain: "db") do
      resources :activities, only: [:index]
      resource :search, only: [:show]

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
        end
      end

      resources :staffs, only: %i(edit update destroy) do
        member do
          get :activities
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

        resource :item, except: [:index]
        resources :casts, only: %i(index new create)
        resources :episodes, only: %i(index new create)
        resources :programs, only: %i(index new create)
        resources :staffs, only: %i(index new create)
      end

      root "home#index"
    end
  end

  resources :settings, only: [:index]
  scope :settings do
    resource :account, only: [:show, :update]
    resource :profile, only: [:show, :update]
    resource :sayonara, only: [:show], controller: :sayonara
    resources :mutes, only: [:index]
    resources :options, only: [:index]
    resources :providers, only: [:index, :destroy]
    resources :apps, only: [:index] do
      patch :revoke
    end

    patch "options", to: "options#update"
  end

  resources :characters, only: %i(show) do
    resources :images, only: %i(index new create destroy), controller: :character_images
  end

  resources :character_images, only: [] do
    resources :reports, only: %i(create), controller: :character_image_reports
  end

  resource :channel, only: [] do
    resources :works, only: [:index], controller: "channel_works"
  end
  resources :channels, only: [:index]

  resources :checkins, only: [] do
    # 旧リダイレクト用URL
    get "redirect/:provider/:url_hash",
      on: :collection,
      as: :redirect,
      to: "checkins#redirect",
      provider: /fb|tw/,
      url_hash: /[0-9a-zA-Z_-]{10}/
  end

  resource :confirmation, only: [:show]
  resource :search, only: [:show]
  resources :friends, only: [:index]
  resources :mute_users, only: [:destroy]
  resources :notifications, only: [:index]
  resources :organizations, only: [:show]
  resources :people, only: [:show]
  resources :programs, only: [:index]

  resources :users, only: [] do
    collection do
      delete :destroy
    end
  end

  scope "@:username", username: /[A-Za-z0-9_]+/ do
    get :following, to: "users#following", as: :following_user
    get :followers, to: "users#followers", as: :followers_user

    get ":status_kind",
      to: "users#works",
      as: :user_works,
      constraints: {
        status_kind: /wanna_watch|watching|watched|on_hold|stop_watching/
      }

    resources :comments, only: %i(edit update destroy)
    resources :records, except: %i(index) do
      resources :comments, only: %i(create)
    end

    root to: "users#show", as: :user
  end

  resources :works, only: [:index, :show] do
    resources :characters, only: %i(index)
    resources :episodes, only: %i(index show)
    resources :images, controller: :work_images, only: %i(index destroy)
    resources :staffs, only: %i(index)

    collection do
      get :popular
      get ":slug",
        action: :season,
        slug: /[0-9]{4}-(all|spring|summer|autumn|winter)/,
        as: :season
      post :switch
    end

    resources :statuses, only: [] do
      post :select, on: :collection
    end
  end

  get "about",   to: "pages#about"
  get "privacy", to: "pages#privacy"
  get "terms",   to: "pages#terms"

  # 新リダイレクト用URL
  get "r/:provider/:url_hash",
    to: "checkins#redirect",
    provider: /fb|tw/,
    url_hash: /[0-9a-zA-Z_-]{10}/

  root "home#index"
end
