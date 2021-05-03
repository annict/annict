# frozen_string_literal: true

module Db
  class CastRowsForm
    include ActiveModel::Model
    include Virtus.model
    include ResourceRows

    row_model Cast

    attr_accessor :work

    validate :valid_resource

    private

    def attrs_list
      casts_count = @work.casts.count
      @attrs_list ||= fetched_rows.map.with_index { |row_data, i|
        {
          work_id: @work.id,
          person_id: row_data[:person][:id],
          character_id: row_data[:character][:id],
          sort_number: (i + casts_count) * 10
        }
      }
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        character = Character.only_kept.where(id: row_columns[0])
          .or(Character.only_kept.where(name: row_columns[0]))
          .order(updated_at: :desc)
          .first
        person = Person.only_kept.where(id: row_columns[1])
          .or(Person.only_kept.where(name: row_columns[1])).first

        {
          character: {id: character&.id, value: row_columns[0]},
          person: {id: person&.id, value: row_columns[1]}
        }
      end
    end

    def valid_resource
      fetched_rows.each do |row_data|
        row_data.slice(:character, :person).each do |_, data|
          next if data[:id].present?
          i18n_path = "activemodel.errors.forms.db/cast_rows_form.invalid"
          errors.add(:rows, I18n.t(i18n_path, value: data[:value]))
        end
      end
    end
  end
end
