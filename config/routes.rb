Annict::Application.routes.draw do
  devise_for :staffs,
    path: :marie,
    controllers: { sessions: 'marie/sessions' },
    path_names: { sign_in: :signin, sign_out: :signout }

  devise_for :users,
              controllers: {
                omniauth_callbacks: :callbacks,
                registrations: :registrations
              },
              path_names: { sign_out: 'signout' }

  namespace :api do
    resources :activities, only: [:index]
    resources :receptions, only: [:create, :destroy]
    resources :users,      only: [] do
      get :activities
    end
    resource  :user,       only: [] do
      resources :programs, only: [:index], controller: 'user_programs'
    end
    resources :works,      only: [] do
      post :hide, on: :member

      resources :channels, only: [] do
        post :select, on: :collection
      end
    end
  end

  namespace :marie do
    resources :works do
      get :on_air, on: :collection

      resources :episodes, only: [:index, :edit, :update, :destroy] do
        collection do
          get  :new_from_csv
          post :create_from_csv
          post :update_sort_number
        end
      end

      resources :items
    end

    root 'home#index'
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

  resources :programs, only: [:index]

  resource :setting, only: [:edit, :update]

  resources :statuses, only: [] do
    delete :like, to: 'likes#status_destroy'
    post   :like, to: 'likes#status_create'
  end

  resources :users,    only: [:show] do
    member do
      delete :unfollow, controller: :follows, action: :destroy
      get ':status_kind',
        to: 'users#works',
        as: :user_works,
        constraints: { status_kind: /wanna_watch|watching|watched|stop_watching/ }
      post   :follow,   controller: :follows, action: :create
    end
  end

  resources :works, only: [:index, :show] do
    collection do
      get :on_air
      get :popular
      get :recommend
      get :season
    end

    resources :appeals,      only: [:create]
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
