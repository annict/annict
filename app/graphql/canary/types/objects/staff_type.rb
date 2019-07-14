# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class StaffType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :annict_id, Integer,
          null: false

        field :name, String,
          null: false

        field :name_en, String,
          null: false

        field :role, String,
          null: false

        field :role_en, String,
          null: false

        field :sort_number, Integer,
          null: false

        field :work, Canary::Types::Objects::WorkType,
          null: false

        field :resource, Canary::Types::Unions::StaffResourceItem,
          null: false

        def role
          return object.role_other if object.role_value == "other"
          I18n.t("enumerize.staff.role.#{object.role_value}", locale: :ja)
        end

        def role_en
          return object.role_other_en if object.role_value == "other"
          I18n.t("enumerize.staff.role.#{object.role_value}", locale: :en)
        end
      end
    end
  end
end
