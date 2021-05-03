# frozen_string_literal: true

FactoryBot.define do
  factory :db_activity do
    association :user, factory: :registered_user

    factory :works_create_activity do
      trackable { create(:work) }
      action { "works.create" }
      parameters { {new: trackable.attributes, old: {}} }
      root_resource { create(:work) }
    end
  end
end
