# frozen_string_literal: true

USERNAME_FORMAT = /[A-Za-z0-9_]+/.freeze

get "dummy_image", to: "application#dummy_image" if Rails.env.test?

devise_for :users,
  controllers: { omniauth_callbacks: :callbacks },
  skip: %i(passwords registrations sessions)

use_doorkeeper do
  controllers applications: "oauth/applications"
  skip_controllers :authorized_applications
end

resource :confirmation, only: [:show]
resource :search, only: [:show]
resource :track, only: :show
resource :work_display_option, only: %i(show)
resources :comments, only: %i(edit update destroy)
resources :faqs, only: %i(index)
resources :friends, only: [:index]
resources :mute_users, only: [:destroy]
resources :notifications, only: [:index]
resources :programs, only: %i(index), controller: :slots
resources :review_comments, only: %i(edit update destroy)
resources :supporters, only: %i(index)

resources :settings, only: [:index]
scope :settings do
  resource :account, only: %i(show update)
  resource :profile, only: %i(show update), as: :profile_setting
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
    to: "episode_records#redirect",
    provider: /fb|tw/,
    url_hash: /[0-9a-zA-Z_-]{10}/
end

resources :episodes, only: [] do
  resources :records, only: %i(create edit update), controller: :episode_records do
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

scope "@:username", username: USERNAME_FORMAT do
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
  resources :reviews, only: %i(show)

  resources :records, only: %i(show destroy) do
    resources :comments, only: %i(create)
  end
end

resources :works, only: %i(index) do
  resources :records, only: %i(index create edit update), controller: :work_records

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
  to: "episode_records#redirect",
  provider: /fb|tw/,
  url_hash: /[0-9a-zA-Z_-]{10}/

root "home#show",
  constraints: Annict::RoutingConstraints::Member.new
root "welcome#show",
  constraints: Annict::RoutingConstraints::Guest.new,
  # Set :as option to avoid two routes with the same name
  as: nil

scope module: :v4 do
  constraints format: "html" do
    devise_scope :user do
      match "/oauth_users",     via: :post,   as: :oauth_users,       to: "oauth_users#create"
      match "/oauth_users/new", via: :get,    as: :new_oauth_user,    to: "oauth_users#new"
      match "/password",        via: :patch,  as: :password,          to: "passwords#update"
      match "/password",        via: :post,                           to: "passwords#create"
      match "/password/edit",   via: :get,    as: :edit_password,     to: "passwords#edit"
      match "/password/new",    via: :get,    as: :new_password,      to: "passwords#new"
      match "/sign_in",         via: :get,    as: :new_user_session,  to: "sessions#new"
      match "/sign_in",         via: :get,    as: :sign_in,           to: "sessions#new"
      match "/sign_in",         via: :post,   as: :user_session,      to: "sessions#create"
      match "/sign_out",        via: :delete, as: :sign_out,          to: "sessions#destroy"
      match "/sign_up",         via: :get,    as: :sign_up,           to: "registrations#new"
      match "/sign_up",         via: :post,   as: :user_registration, to: "registrations#create"
    end

    match "/@:username",         via: :get,   as: :profile_detail, to: "users#show",    username: USERNAME_FORMAT
    match "/@:username/records", via: :get,   as: :record_list,    to: "records#index", username: USERNAME_FORMAT
    match "/timeline_mode",      via: :patch, as: :timeline_mode,  to: "timeline_mode#update"
    match "/works/:id",          via: :get,   as: :work,           to: "works#show"
  end
end
