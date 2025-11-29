# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class PersonFavoriteType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        field :user, Canary::Types::Objects::UserType, null: false
        field :person, Canary::Types::Objects::PersonType, null: false
        field :watched_work_count, Integer, null: false
        field :created_at, Canary::Types::Scalars::DateTime, null: false

        def user
          Canary::RecordLoader.for(User).load(object.user_id)
        end

        def person
          Canary::RecordLoader.for(Person).load(object.person_id)
        end

        def watched_work_count
          object.watched_works_count
        end
      end
    end
  end
end
