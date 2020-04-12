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

  resources :comments, only: %i(create destroy)

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

  resources :trailers, only: %i(edit update destroy) do
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

  resources :works, only: [] do
    resources :trailers, only: %i(index new create)
  end

  root "home#index"
end

scope module: :db do
  constraints format: "html" do
    # rubocop:disable Layout/ExtraSpacing, Layout/LineLength
    match "/db/activities",                         via: :get,    as: :db_activity_list,          to: "activities#index"
    match "/db/casts/:id",                          via: :delete, as: :db_cast_detail,            to: "casts#destroy"
    match "/db/casts/:id",                          via: :patch,                                  to: "casts#update"
    match "/db/casts/:id/edit",                     via: :get,    as: :db_edit_cast,              to: "casts#edit"
    match "/db/casts/:id/publishing",               via: :delete, as: :db_cast_publishing,        to: "cast_publishings#destroy"
    match "/db/casts/:id/publishing",               via: :post,                                   to: "cast_publishings#create"
    match "/db/episodes/:id",                       via: :delete, as: :db_episode_detail,         to: "episodes#destroy"
    match "/db/episodes/:id",                       via: :patch,                                  to: "episodes#update"
    match "/db/episodes/:id/edit",                  via: :get,    as: :db_edit_episode,           to: "episodes#edit"
    match "/db/episodes/:id/publishing",            via: :delete, as: :db_episode_publishing,     to: "episode_publishings#destroy"
    match "/db/episodes/:id/publishing",            via: :post,                                   to: "episode_publishings#create"
    match "/db/programs/:id",                       via: :delete, as: :db_program_detail,         to: "programs#destroy"
    match "/db/programs/:id",                       via: :patch,                                  to: "programs#update"
    match "/db/programs/:id/edit",                  via: :get,    as: :db_edit_program,           to: "programs#edit"
    match "/db/programs/:id/publishing",            via: :delete, as: :db_program_publishing,     to: "program_publishings#destroy"
    match "/db/programs/:id/publishing",            via: :post,                                   to: "program_publishings#create"
    match "/db/series",                             via: :get,    as: :db_series_list,            to: "series#index"
    match "/db/series",                             via: :post,                                   to: "series#create"
    match "/db/series/:id",                         via: :delete, as: :db_series_detail,          to: "series#destroy"
    match "/db/series/:id",                         via: :patch,                                  to: "series#update"
    match "/db/series/:id/edit",                    via: :get,    as: :db_edit_series,            to: "series#edit"
    match "/db/series/:id/publishing",              via: :delete, as: :db_series_publishing,      to: "series_publishings#destroy"
    match "/db/series/:id/publishing",              via: :post,                                   to: "series_publishings#create"
    match "/db/series/:series_id/series_works",     via: :get,    as: :db_series_work_list,       to: "series_works#index"
    match "/db/series/:series_id/series_works",     via: :post,                                   to: "series_works#create"
    match "/db/series/:series_id/series_works/new", via: :get,    as: :db_new_series_work,        to: "series_works#new"
    match "/db/series/new",                         via: :get,    as: :db_new_series,             to: "series#new"
    match "/db/series_works/:id",                   via: :delete, as: :db_series_work_detail,     to: "series_works#destroy"
    match "/db/series_works/:id",                   via: :patch,                                  to: "series_works#update"
    match "/db/series_works/:id/edit",              via: :get,    as: :db_edit_series_work,       to: "series_works#edit"
    match "/db/series_works/:id/publishing",        via: :delete, as: :db_series_work_publishing, to: "series_work_publishings#destroy"
    match "/db/series_works/:id/publishing",        via: :post,                                   to: "series_work_publishings#create"
    match "/db/slots/:id",                          via: :delete, as: :db_slot_detail,            to: "slots#destroy"
    match "/db/slots/:id",                          via: :patch,                                  to: "slots#update"
    match "/db/slots/:id/edit",                     via: :get,    as: :db_edit_slot,              to: "slots#edit"
    match "/db/slots/:id/publishing",               via: :delete, as: :db_slot_publishing,        to: "slot_publishings#destroy"
    match "/db/slots/:id/publishing",               via: :post,                                   to: "slot_publishings#create"
    match "/db/staffs/:id",                         via: :delete, as: :db_staff_detail,           to: "staffs#destroy"
    match "/db/staffs/:id",                         via: :patch,                                  to: "staffs#update"
    match "/db/staffs/:id/edit",                    via: :get,    as: :db_edit_staff,             to: "staffs#edit"
    match "/db/staffs/:id/publishing",              via: :delete, as: :db_staff_publishing,       to: "staff_publishings#destroy"
    match "/db/staffs/:id/publishing",              via: :post,                                   to: "staff_publishings#create"
    match "/db/works",                              via: :get,    as: :db_work_list,              to: "works#index"
    match "/db/works",                              via: :post,                                   to: "works#create"
    match "/db/works/:id",                          via: :delete, as: :db_work_detail,            to: "works#destroy"
    match "/db/works/:id",                          via: :patch,                                  to: "works#update"
    match "/db/works/:id/edit",                     via: :get,    as: :db_edit_work,              to: "works#edit"
    match "/db/works/:id/publishing",               via: :delete, as: :db_work_publishing,        to: "work_publishings#destroy"
    match "/db/works/:id/publishing",               via: :post,                                   to: "work_publishings#create"
    match "/db/works/:work_id/casts",               via: :get,    as: :db_cast_list,              to: "casts#index"
    match "/db/works/:work_id/casts",               via: :post,                                   to: "casts#create"
    match "/db/works/:work_id/casts/new",           via: :get,    as: :db_new_cast,               to: "casts#new"
    match "/db/works/:work_id/episodes",            via: :get,    as: :db_episode_list,           to: "episodes#index"
    match "/db/works/:work_id/episodes",            via: :post,                                   to: "episodes#create"
    match "/db/works/:work_id/episodes/new",        via: :get,    as: :db_new_episode,            to: "episodes#new"
    match "/db/works/:work_id/image",               via: :get,    as: :db_work_image_detail,      to: "work_images#show"
    match "/db/works/:work_id/image",               via: :patch,                                  to: "work_images#update"
    match "/db/works/:work_id/image",               via: :post,                                   to: "work_images#create"
    match "/db/works/:work_id/programs",            via: :get,    as: :db_program_list,           to: "programs#index"
    match "/db/works/:work_id/programs",            via: :post,                                   to: "programs#create"
    match "/db/works/:work_id/programs/new",        via: :get,    as: :db_new_program,            to: "programs#new"
    match "/db/works/:work_id/slots",               via: :get,    as: :db_slot_list,              to: "slots#index"
    match "/db/works/:work_id/slots",               via: :post,                                   to: "slots#create"
    match "/db/works/:work_id/slots/new",           via: :get,    as: :db_new_slot,               to: "slots#new"
    match "/db/works/:work_id/staffs",              via: :get,    as: :db_staff_list,             to: "staffs#index"
    match "/db/works/:work_id/staffs",              via: :post,                                   to: "staffs#create"
    match "/db/works/:work_id/staffs/new",          via: :get,    as: :db_new_staff,              to: "staffs#new"
    match "/db/works/new",                          via: :get,    as: :db_new_work,               to: "works#new"
    # rubocop:enable Layout/ExtraSpacing, Layout/LineLength
  end
end
