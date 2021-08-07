# frozen_string_literal: true

FactoryBot.define do
  factory :db_activity do
    association :user, factory: :registered_user

    factory :animes_create_activity do
      trackable { create(:anime) }
      action { "works.create" }
      parameters { {new: trackable.attributes, old: {}} }
      root_resource { create(:anime) }
    end
  end
end
