# frozen_string_literal: true

USERNAME_FORMAT = /[A-Za-z0-9_]+/

get "dummy_image", to: "application#dummy_image" if Rails.env.test?

devise_for :users,
  controllers: {omniauth_callbacks: :callbacks},
  skip: %i[passwords registrations sessions]

use_doorkeeper do
  controllers applications: "oauth/applications"
  skip_controllers :authorized_applications
end

resource :confirmation, only: [:show]
resource :search, only: [:show]
resources :comments, only: %i[edit update destroy]
resources :faqs, only: %i[index]
resources :friends, only: [:index]
resources :mute_users, only: [:destroy]
resources :notifications, only: [:index]
resources :review_comments, only: %i[edit update destroy]
resources :supporters, only: %i[index]

resources :settings, only: [:index]
scope :settings do
  resource :account, only: %i[show update]
  # TODO: /my/profile パスに変更する
  resource :profile, only: %i[show update], as: :profile_setting
  resources :mutes, only: [:index]
  resources :options, only: [:index]
  resources :providers, only: %i[index destroy]

  patch "options", to: "options#update"
end
namespace :settings do
  resource :password, only: %i[update]

  resources :apps, only: %i[index] do
    patch :revoke
  end

  resource :email_notification, only: %i[show update] do
    get :unsubscribe, on: :collection
  end

  resources :tokens, only: %i[new create edit update destroy]
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
  resources :records, only: [], controller: :episode_records do
    post :switch, on: :collection
  end
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

  resources :tags, only: %i[show], controller: :user_work_tags, as: :user_work_tag
  resources :reviews, only: %i[show]

  resources :records, only: %i[] do
    resources :comments, only: %i[create]
  end
end

resources :works, only: %i[index] do
  resources :episodes, only: [] do
    resources :checkins, only: %i[show]
  end
end

# 新リダイレクト用URL
get "r/:provider/:url_hash",
  to: "episode_records#redirect",
  provider: /fb|tw/,
  url_hash: /[0-9a-zA-Z_-]{10}/

devise_scope :user do
  match "/legacy/sign_in", via: :get, as: :legacy_sign_in, to: "v6/legacy/sessions#new"
  match "/legacy/sign_in", via: :post, as: :user_session, to: "v6/legacy/sessions#create"
  match "/sign_out", via: :delete, as: :sign_out, to: "devise/sessions#destroy"
  match "/user_email", via: :patch, as: :user_email, to: "v4/user_emails#update"
  match "/user_email/callback", via: :get, as: :user_email_callback, to: "v4/user_email_callbacks#show"
end

match "/@:username", via: :get, as: :profile, to: "v6/users#show", username: USERNAME_FORMAT
match "/@:username/:status_kind", via: :get, as: :library, to: "v4/libraries#show", username: USERNAME_FORMAT, status_kind: /wanna_watch|watching|watched|on_hold|stop_watching/
match "/@:username/favorite_characters", via: :get, as: :favorite_character_list, to: "v3/favorite_characters#index", username: USERNAME_FORMAT
match "/@:username/favorite_organizations", via: :get, as: :favorite_organization_list, to: "v3/favorite_organizations#index", username: USERNAME_FORMAT
match "/@:username/favorite_people", via: :get, as: :favorite_person_list, to: "v3/favorite_people#index", username: USERNAME_FORMAT
match "/@:username/records", via: :get, as: :record_list, to: "v4/records#index", username: USERNAME_FORMAT
match "/@:username/records/:record_id", via: :delete, to: "v4/records#destroy", username: USERNAME_FORMAT
match "/@:username/records/:record_id", via: :get, to: "v4/records#show", username: USERNAME_FORMAT
match "/@:username/records/:record_id", via: :patch, as: :record, to: "records#update", username: USERNAME_FORMAT
match "/@:username/records/:record_id", via: :patch, to: "v4/records#update", username: USERNAME_FORMAT
match "/characters/:character_id", via: :get, as: :character, to: "v3/characters#show"
match "/characters/:character_id/fans", via: :get, as: :character_fan_list, to: "v3/character_fans#index"
match "/episode_records", via: :patch, as: :episode_record_mutation, to: "v4/episode_records#update"
match "/episodes/:episode_id/records", via: :post, as: :episode_record_list, to: "episode_records#create"
match "/fragment/@:username/records/:record_id/edit", via: :get, as: :fragment_edit_record, to: "v6/fragment/records#edit", username: USERNAME_FORMAT
match "/fragment/activity_groups/:activity_group_id/items", via: :get, as: :fragment_activity_item_list, to: "v6/fragment/activity_items#index"
match "/fragment/episodes/:episode_id/records", via: :get, as: :fragment_episode_record_list, to: "v6/fragment/episode_records#index"
match "/fragment/receive_channel_buttons", via: :get, as: :fragment_receive_channel_button_list, to: "v6/fragment/receive_channel_buttons#index"
match "/fragment/trackable_anime/:anime_id", via: :get, as: :fragment_trackable_anime, to: "v6/fragment/trackable_anime#show"
match "/fragment/trackable_episodes", via: :get, as: :fragment_trackable_episode_list, to: "v6/fragment/trackable_episodes#index"
match "/fragment/trackable_episodes/:episode_id", via: :get, as: :fragment_trackable_episode, to: "v6/fragment/trackable_episodes#show"
match "/legal", via: :get, as: :legal, to: "v6/pages#legal"
match "/my/profile", via: :get, as: :my_profile, to: "my/profiles#show"
match "/organizations/:organization_id", via: :get, as: :organization, to: "v3/organizations#show"
match "/organizations/:organization_id/fans", via: :get, as: :organization_fan_list, to: "v3/organization_fans#index"
match "/people/:person_id", via: :get, as: :person, to: "v3/people#show"
match "/people/:person_id/fans", via: :get, as: :person_fan_list, to: "v3/person_fans#index"
match "/privacy", via: :get, as: :privacy, to: "v6/pages#privacy"
match "/registrations/new", via: :get, as: :new_registration, to: "v6/registrations#new"
match "/sign_in", via: :get, as: :new_user_session, to: "v6/sign_in#new" # for Devise
match "/sign_in", via: :get, as: :sign_in, to: "v6/sign_in#new"
match "/sign_in/callback", via: :get, as: :sign_in_callback, to: "v6/sign_in_callbacks#show"
match "/sign_up", via: :get, as: :sign_up, to: "v6/sign_up#new"
match "/terms", via: :get, as: :terms, to: "v6/pages#terms"
match "/track", via: :get, as: :track, to: "tracks#show"
match "/work_display_option", via: :get, as: :work_display_option, to: "v3/work_display_options#show"
match "/works/:anime_id", via: :get, as: :anime, to: "v4/works#show"
match "/works/:anime_id/episodes", via: :get, as: :episode_list, to: "v4/episodes#index"
match "/works/:anime_id/episodes/:episode_id", via: :get, as: :episode, to: "v6/episodes#show"
match "/works/:anime_id/records", via: :get, as: :anime_record_list, to: "v4/anime_records#index"
match "/works/:anime_id/records", via: :post, to: "v4/anime_records#create"
match "/works/:slug", via: :get, as: :seasonal_anime_list, to: "v3/works#season", slug: /[0-9]{4}-(all|spring|summer|autumn|winter)/
match "/works/newest", via: :get, as: :newest_anime_list, to: "v3/works#newest"
match "/works/popular", via: :get, as: :popular_anime_list, to: "v3/works#popular"

root "v6/home#show",
  constraints: Annict::RoutingConstraints::Member.new
root "v6/welcome#show",
  constraints: Annict::RoutingConstraints::Guest.new,
  # Set :as option to avoid two routes with the same name
  as: nil
