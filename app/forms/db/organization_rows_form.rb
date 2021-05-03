# frozen_string_literal: true

module Db
  class OrganizationRowsForm
    include ActiveModel::Model
    include Virtus.model
    include ResourceRows

    row_model Organization

    attribute :rows, String

    validates :rows, presence: true
    validate :valid_organization

    def save!
      new_resources_with_user.each(&:save_and_create_activity!)
    end

    private

    def attrs_list
      @attrs_list ||= fetched_rows.map { |row_data|
        {
          name: row_data[:name][:value],
          name_kana: row_data[:name_kana][:value].presence || ""
        }
      }
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        {
          name: {value: row_columns[0]},
          name_kana: {value: row_columns[1]}
        }
      end
    end

    def valid_organization
      return if new_resources.all?(&:valid?)

      new_resources.each do |c|
        next if c.valid?

        message = "\"#{c.name}\"#{c.errors.messages[:name].first}"
        errors.add(:rows, message)
      end
    end
  end
end
