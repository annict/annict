# frozen_string_literal: true

module Db
  class CharacterRowsForm
    include ActiveModel::Model
    include Virtus.model
    include ResourceRows

    row_model Character

    attribute :rows, String

    validates :rows, presence: true
    validate :valid_character

    def save!
      new_resources_with_user.each(&:save_and_create_activity!)
    end

    private

    def attrs_list
      @attrs_list ||= fetched_rows.map { |row_data|
        {
          name: row_data[:name][:value],
          name_kana: row_data[:name_kana][:value].presence || "",
          series_id: row_data[:series][:id]
        }
      }
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        series = Series.only_kept.where(id: row_columns[2])
          .or(Series.only_kept.where(name: row_columns[2])).first

        {
          name: {value: row_columns[0]},
          name_kana: {value: row_columns[1]},
          series: {id: series&.id, value: row_columns[2]}
        }
      end
    end

    def valid_character
      return if new_resources.all?(&:valid?)

      new_resources.each do |c|
        next if c.valid?

        message = "\"#{c.name}\": #{c.errors.full_messages.first}"
        errors.add(:rows, message)
      end
    end
  end
end
