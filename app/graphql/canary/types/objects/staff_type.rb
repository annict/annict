# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class StaffType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :database_id, Integer,
          null: false

        field :name, String,
          null: false

        field :name_en, String,
          null: false

        field :accurate_name, String,
          null: false,
          description: "担当者名。名義が異なる場合2つの名前を併記する。例: ふでやすかずゆき (筆安一幸)"

        field :accurate_name_en, String,
          null: false,
          description: "担当者名 (英)。名義が異なる場合2つの名前を併記する。"

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

        def accurate_name
          object.decorate.accurate_name
        end

        def accurate_name_en
          object.decorate.accurate_name_en
        end

        def role
          return object.role_other if object.role_value == "other"

          I18n.t("enumerize.staff.role.#{object.role_value}", locale: :ja)
        end

        def role_en
          return object.role_other_en if object.role_value == "other"

          I18n.t("enumerize.staff.role.#{object.role_value}", locale: :en)
        end

        def work
          RecordLoader.for(Work).load(object.work_id)
        end
      end
    end
  end
end
