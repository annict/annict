# frozen_string_literal: true

Rails.application.routes.draw do
  if Rails.env.development?
    mount LetterOpenerWeb::Engine, at: "/low"
    mount Dmmyix::Engine, at: "/dmmyix"
    mount GraphiQL::Rails::Engine, at: "/graphiql", graphql_path: "/graphql"
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
      resource :impression, only: %i(show update)
      resource :programs_sort_type, only: [:update]
      resource :records_sort_type, only: %i(update)
      resource :search, only: [:show]
      resources :activities, only: [:index]
      resources :characters, only: [:index]
      resources :items, only: %i(create)
      resources :mute_users, only: [:create]
      resources :organizations, only: [:index]
      resources :people, only: [:index]
      resources :receptions, only: %i(create destroy)
      resources :series_list, only: %i(index)
      resources :works, only: %i(index show)

      resource :amazon, only: [], controller: :amazon do
        get :search
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

      resources :records, only: %i(create) do
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

  scope module: :api do
    constraints(subdomain: "api") do
      namespace :v1 do
        resources :activities, only: %i(index)
        resources :episodes, only: [:index]
        resources :followers, only: %i(index)
        resources :following, only: %i(index)
        resources :records, only: [:index]
        resources :reviews, only: %i(index)
        resources :users, only: %i(index)
        resources :works, only: [:index]

        namespace :me do
          resources :following_activities, only: %i(index)
          resources :programs, only: [:index]
          resources :records, only: %i(create update destroy)
          resources :reviews, only: %i(create update destroy)
          resources :statuses, only: [:create]
          resources :works, only: [:index]

          root "index#show"
        end

        match "*path", to: "application#not_found", via: :all
      end
    end
  end

  namespace :db do
    resources :activities, only: [:index]
    resources :channels, only: [:index]
    resource :search, only: [:show]

    resources :characters, except: [:show] do
      member do
        get :activities
        patch :hide
      end
      resource :image, controller: :character_images, only: %i(show create update destroy)
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

    resources :pvs, only: %i(edit update destroy) do
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
      resources :pvs, only: %i(index new create)
      resources :staffs, only: %i(index new create)
      resources :program_details, only: %i(index new create)
    end

    root "home#index"
  end

  namespace :forum do
    resources :categories, only: %i(show)

    resources :posts, only: %i(new create show edit update) do
      resources :comments, only: %i(new create edit update)
    end

    root "home#index"
  end

  namespace :userland do
    resources :projects, except: %i(index)

    root "home#index"
  end

  resource :confirmation, only: [:show]
  resource :menu, only: %i(show)
  resource :search, only: [:show]
  resource :track, only: %i(show)
  resources :comments, only: %i(edit update destroy)
  resources :faqs, only: %i(index)
  resources :friends, only: [:index]
  resources :mute_users, only: [:destroy]
  resources :notifications, only: [:index]
  resources :programs, only: [:index]
  resources :review_comments, only: %i(edit update destroy)

  resources :settings, only: [:index]
  scope :settings do
    resource :account, only: %i(show update)
    resource :profile, only: %i(show update)
    resources :mutes, only: [:index]
    resources :options, only: [:index]
    resources :providers, only: %i(index destroy)

    patch "options", to: "options#update"
  end
  namespace :settings do
    resource :password, only: %i(update)

    resources :apps, only: %i(index) do
      patch :revoke
    end

    resource :email_notification, only: %i(show update) do
      get :unsubscribe, on: :collection
    end

    resources :tokens, only: %i(new create edit update destroy)
  end

  resources :characters, only: %i(show) do
    resources :fans, only: %i(index), controller: "character_fans"
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

  resources :episodes, only: [] do
    resources :items, only: %i(new destroy), controller: :episode_items
    resources :records, only: [] do
      post :switch, on: :collection
    end
  end

  resources :organizations, only: %i(show) do
    resources :fans, only: %i(index), controller: "organization_fans"
  end

  resources :people, only: %i(show) do
    resources :fans, only: %i(index), controller: "person_fans"
  end

  resources :users, only: [] do
    collection do
      delete :destroy
    end
  end

  scope "@:username", username: /[A-Za-z0-9_]+/ do
    get :following, to: "users#following", as: :following_user
    get :followers, to: "users#followers", as: :followers_user
    get :ics, to: "ics#show", as: :user_ics

    get ":status_kind",
      to: "libraries#show",
      as: :library,
      constraints: {
        status_kind: /wanna_watch|watching|watched|on_hold|stop_watching/
      }

    resources :favorite_characters, only: %i(index)
    resources :favorite_organizations, only: %i(index)
    resources :favorite_people, only: %i(index)
    resources :tags, only: %i(show), controller: :user_work_tags, as: :user_work_tag

    resources :records, only: %i(create show edit update destroy) do
      resources :comments, only: %i(create)
    end

    resources :reviews, only: %i(index show) do
      resources :review_comments, only: %i(create)
    end

    root to: "users#show", as: :user
  end

  resources :works, only: %i(index show) do
    resources :characters, only: %i(index)
    resources :items, only: %i(index new destroy), controller: :work_items
    resources :staffs, only: %i(index)
    resources :reviews, only: %i(new create edit update destroy)
    resources :reviews, only: %i(index), controller: :work_reviews

    resources :episodes, only: %i(index show) do
      resources :checkins, only: %i(show)
    end

    collection do
      get :newest
      get :popular
      get ":slug",
        action: :season,
        slug: /[0-9]{4}-(all|spring|summer|autumn|winter)/,
        as: :season
      post :switch
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

  post "/graphql", to: "graphql#execute"

  root "home#index"
end
