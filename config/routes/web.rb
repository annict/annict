# frozen_string_literal: true

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

resource :confirmation, only: [:show]
resource :menu, only: %i(show)
resource :search, only: [:show]
resource :work_display_option, only: %i(show)
resources :activities, only: %i(index)
resources :comments, only: %i(edit update destroy)
resources :faqs, only: %i(index)
resources :friends, only: [:index]
resources :mute_users, only: [:destroy]
resources :notifications, only: [:index]
resources :programs, only: [:index]
resources :review_comments, only: %i(edit update destroy)
resources :supporters, only: %i(index)

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
    to: "records#redirect",
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
  resources :items, only: %i(index new destroy), controller: :work_items
  resources :reviews, only: %i(create edit update destroy)
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
  end
end

get "about", to: "pages#about"
get "legal", to: "pages#legal"
get "privacy", to: "pages#privacy"
get "terms", to: "pages#terms"

# 新リダイレクト用URL
get "r/:provider/:url_hash",
  to: "records#redirect",
  provider: /fb|tw/,
  url_hash: /[0-9a-zA-Z_-]{10}/

root "home#index"
