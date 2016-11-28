# frozen_string_literal: true

module DB
  class CastRowsForm
    include ActiveModel::Model
    include Virtus.model
    include ResourceRows

    row_model Cast

    attr_accessor :work

    validate :valid_resource

    private

    def attrs_list
      @attrs_list ||= fetched_rows.map do |row_data|
        {
          work_id: @work.id,
          person_id: row_data[:person][:id],
          character_id: row_data[:character][:id],
          part: ""
        }
      end
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        character = Character.where(id: row_columns[0]).
          or(Character.where(name: row_columns[0])).first
        person = Person.where(id: row_columns[1]).
          or(Person.where(name: row_columns[1])).first

        {
          character: { id: character&.id, value: row_columns[0] },
          person: { id: person&.id, value: row_columns[1] }
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
