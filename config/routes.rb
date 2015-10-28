Annict::Application.routes.draw do
  if Rails.env.development?
    mount LetterOpenerWeb::Engine, at: '/low'
  end

  if Rails.env.test?
    # テスト実行時にDragonflyでアップロードした画像を読み込むときに呼ばれるアクション
    get ':image_size/:image_path',
        to: 'application#dummy_image',
        image_size: /[0-9]+x[0-9]+e*/,
        image_path: %r([0-9]{4}/[0-9]{2}/[0-9]{2}/.+)
  end

  devise_for :users,
              controllers: { omniauth_callbacks: :callbacks },
              skip: [:registrations],
              path_names: { sign_out: 'signout' }

  devise_scope :user do
    get 'users/sign_up', to: 'registrations#new', as: :new_user_registration
    post 'users', to: 'registrations#create', as: :user_registration
  end

  use_doorkeeper

  namespace :api do
    namespace :private do
      resources :activities, only: [:index]
      resources :receptions, only: [:create, :destroy]
      resources :tips, only: [] do
        post :finish, on: :collection
      end
      resources :users, only: [] do
        get :activities
      end
      resource :user, only: [] do
        resources :checks, only: [:index], controller: "user_checks" do
          patch :skip_episode
        end
        resources :programs, only: [:index], controller: 'user_programs'
      end
      resources :works, only: [] do
        resources :channels, only: [] do
          post :select, on: :collection
        end
      end
    end
  end

  scope module: :api do
    constraints Annict::Subdomain do
      namespace :v1 do
        resources :users, only: [] do
          get :me, on: :collection
        end
        resources :works, only: [:index]
      end
    end
  end

  namespace :db do
    resources :activities, only: [:index]
    resources :draft_works, only: [:new, :create, :edit, :update]
    resources :edit_requests, only: [:index, :show] do
      member do
        post :close
        post :publish
      end
      resources :comments, only: [:create], controller: "edit_request_comments"
    end
    resources :works, except: [:show] do
      collection do
        get :season
        get :resourceless
        get :search
      end
      member do
        patch :hide
      end
      resources :draft_episodes, only: [:new, :create, :edit, :update]
      resources :draft_items, only: [:new, :create, :edit, :update]
      resources :draft_multiple_episodes, only: [:new, :create, :edit, :update]
      resources :draft_programs, only: [:new, :create, :edit, :update]
      resources :episodes, only: [:index, :edit, :update, :destroy] do
        member do
          patch :hide
        end
      end
      resources :multiple_episodes, only: [:new, :create]
      resources :programs, except: [:show]
      resource :item, except: [:index]
    end

    root "home#index"
  end

  resource  :channel, only: [] do
    resources :works, only: [:index], controller: 'channel_works'
  end
  resources :channels, only: [:index]

  resources :checkins, only: [] do
    # 旧リダイレクト用URL
    get 'redirect/:provider/:url_hash',
      on: :collection,
      as: :redirect,
      to: 'checkins#redirect',
      provider: /fb|tw/,
      url_hash: /[0-9a-zA-Z_-]{10}/

    delete :like, to: 'likes#checkin_destroy'
    post   :like, to: 'likes#checkin_create'
  end

  resources :comments, only: [] do
    delete :like, to: 'likes#comment_destroy'
    post   :like, to: 'likes#comment_create'
  end

  resource :confirmation, only: [:show]

  resources :friends, only: [:index]

  resources :notifications, only: [:index]

  resource  :profile, only: [:update]

  resources :programs, only: [:index]

  resources :providers, only: [:destroy]

  resource :setting, only: [:show, :update]

  resources :statuses, only: [] do
    delete :like, to: 'likes#status_destroy'
    post   :like, to: 'likes#status_create'
  end

  resources :users, only: [:show] do
    collection do
      delete :destroy
      patch :update
      post :share
    end

    member do
      delete :unfollow, controller: :follows, action: :destroy
      get ':status_kind',
        to: 'users#works',
        as: :user_works,
        constraints: { status_kind: /wanna_watch|watching|watched|on_hold|stop_watching/ }
      get :following
      get :followers
      post :follow, controller: :follows, action: :create
    end
  end

  resources :works, only: [:index, :show] do
    collection do
      get :popular
      get ':name',
        action: :season,
        name: /[0-9]{4}-(spring|summer|autumn|winter)/,
        as: :season
    end

    resources :episodes,     only: [:show] do
      resources :checkins do
        resources :comments, only: [:create]
      end
    end

    resources :statuses,     only: [] do
      post :select, on: :collection
    end

    resources :checkins, only: [] do
      post :create_all, on: :collection
    end

    get :search, on: :collection
  end

  get 'about',   to: 'pages#about'
  get 'privacy', to: 'pages#privacy'
  get 'terms',   to: 'pages#terms'

  # 新リダイレクト用URL
  get 'r/:provider/:url_hash',
    to: 'checkins#redirect',
    provider: /fb|tw/,
    url_hash: /[0-9a-zA-Z_-]{10}/

  root 'home#index'
end
