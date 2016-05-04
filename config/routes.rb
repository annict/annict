# frozen_string_literal: true

Annict::Application.routes.draw do
  mount LetterOpenerWeb::Engine, at: "/low" if Rails.env.development?

  get "dummy_image", to: "application#dummy_image" if Rails.env.test?

  devise_for :users,
    controllers: { omniauth_callbacks: :callbacks },
    skip: [:registrations, :sessions]

  devise_scope :user do
    get "sign_up", to: "registrations#new", as: :new_user_registration
    get "sign_in", to: "sessions#new", as: :new_user_session
    post "sign_in", to: "sessions#create", as: :user_session
    post "users", to: "registrations#create", as: :user_registration
    delete "sign_out", to: "sessions#destroy", as: :destroy_user_session
  end

  use_doorkeeper

  namespace :api do
    namespace :internal do
      resource :search, only: [:show]
      resources :activities, only: [:index]
      resources :organizations, only: [:index]
      resources :people, only: [:index]
      resources :receptions, only: [:create, :destroy]

      resources :comments, only: [] do
        delete :like, to: "likes#comment_destroy"
        post   :like, to: "likes#comment_create"
      end

      resources :latest_statuses, only: [:index] do
        patch :skip_episode
      end

      resources :multiple_records, only: [] do
        delete :like, to: "likes#multiple_record_destroy"
        post   :like, to: "likes#multiple_record_create"
      end

      resources :records, only: [] do
        delete :like, to: "likes#record_destroy"
        post   :like, to: "likes#record_create"
      end

      resources :statuses, only: [] do
        delete :like, to: "likes#status_destroy"
        post   :like, to: "likes#status_create"
      end

      resources :tips, only: [] do
        post :finish, on: :collection
      end

      resource :user, only: [] do
        resources :programs, only: [:index], controller: "user_programs"
      end

      resources :works, only: [] do
        get :friends

        resources :channels, only: [] do
          post :select, on: :collection
        end
      end
    end
  end

  scope module: :api do
    constraints Annict::Subdomain do
      namespace :v1 do
        resources :episodes, only: [:index]
        resources :works, only: [:index]

        resource :me, only: [] do
          resources :statuses, only: [:create]
        end

        match "*path", to: "application#not_found", via: :all
      end
    end
  end

  namespace :db do
    resources :activities, only: [:index]
    resources :draft_organizations, only: [:new, :create, :edit, :update]
    resources :draft_people, only: [:new, :create, :edit, :update]
    resources :draft_works, only: [:new, :create, :edit, :update]
    resource :search, only: [:show]

    resources :edit_requests, only: [:index, :show] do
      member do
        post :close
        post :publish
      end
      resources :comments,
        only: [:create, :edit, :update, :destroy],
        controller: "edit_request_comments"
    end

    resources :organizations, except: [:show] do
      patch :hide, on: :member
    end

    resources :people, except: [:show] do
      patch :hide, on: :member
    end

    resources :works, except: [:show] do
      collection do
        get :season
        get :resourceless
      end

      member do
        patch :hide
      end

      resources :draft_casts, only: [:new, :create, :edit, :update]
      resources :draft_episodes, only: [:new, :create, :edit, :update]
      resources :draft_items, only: [:new, :create, :edit, :update]
      resources :draft_multiple_episodes, only: [:new, :create, :edit, :update]
      resources :draft_programs, only: [:new, :create, :edit, :update]
      resources :draft_staffs, only: [:new, :create, :edit, :update]
      resources :multiple_episodes, only: [:new, :create]
      resources :programs, except: [:show]
      resource :item, except: [:index]

      resources :casts, except: [:show] do
        patch :hide, on: :member
      end

      resources :episodes, only: [:index, :edit, :update, :destroy] do
        member do
          patch :hide
        end
      end

      resources :staffs, except: [:show] do
        patch :hide, on: :member
      end
    end

    root "home#index"
  end

  resource :search, only: [:show]

  resources :organizations, only: [:show]
  resources :people, only: [:show]

  resources :settings, only: [:index]
  scope :settings do
    resource :account, only: [:show, :update]
    resource :profile, only: [:show, :update]
    resource :sayonara, only: [:show], controller: :sayonara
    resources :options, only: [:index]
    resources :providers, only: [:index, :destroy]

    patch "options", to: "options#update"
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
  resources :friends, only: [:index]
  resources :notifications, only: [:index]
  resources :programs, only: [:index]

  get "@:username", to: "users#show", username: /[A-Za-z0-9_]+/, as: :user
  get "@:username/:status_kind",
    to: "users#works",
    as: :user_works,
    constraints: {
      username: /[A-Za-z0-9_]+/,
      status_kind: /wanna_watch|watching|watched|on_hold|stop_watching/
    }
  get "@:username/following",
    to: "users#following",
    username: /[A-Za-z0-9_]+/,
    as: :following_user
  get "@:username/followers",
    to: "users#followers",
    username: /[A-Za-z0-9_]+/,
    as: :followers_user
  resources :users, only: [] do
    collection do
      delete :destroy
    end

    member do
      delete :unfollow, controller: :follows, action: :destroy
      post :follow, controller: :follows, action: :create
    end
  end

  resources :works, only: [:index, :show] do
    collection do
      get :popular
      get ":slug",
        action: :season,
        slug: /[0-9]{4}-(all|spring|summer|autumn|winter)/,
        as: :season
    end

    resources :episodes, only: [:show] do
      resources :checkins do
        resources :comments, only: [:create]
      end
    end

    resources :statuses, only: [] do
      post :select, on: :collection
    end

    resources :checkins, only: [] do
      post :create_all, on: :collection
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
