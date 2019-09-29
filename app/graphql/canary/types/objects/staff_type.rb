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

        field :local_accurated_name, String,
          null: false,
          description: "担当者名。名義が異なる場合2つの名前を併記する。例: ふでやすかずゆき (筆安一幸)"

        field :role, String,
          null: false

        field :role_en, String,
          null: false

        field :local_role, String,
          null: false

        field :sort_number, Integer,
          null: false

        field :work, Canary::Types::Objects::WorkType,
          null: false

        field :resource, Canary::Types::Unions::StaffResourceItem,
          null: false

        def local_accurated_name
          object.decorate.local_name_with_old
        end

        def local_role
          object.decorate.role_name
        end

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
