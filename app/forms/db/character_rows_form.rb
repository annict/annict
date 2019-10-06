# frozen_string_literal: true

module DB
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
      @attrs_list ||= fetched_rows.map do |row_data|
        {
          name: row_data[:name][:value],
          series_id: row_data[:series][:id]
        }
      end
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        series = Series.published.where(id: row_columns[1]).
          or(Series.published.where(name: row_columns[1])).first

        {
          name: { value: row_columns[0] },
          series: { id: series&.id, value: row_columns[1] }
        }
      end
    end

    def valid_character
      return if new_resources.all?(&:valid?)

      new_resources.each do |c|
        next if c.valid?
        message = "\"#{c.name}\"#{c.errors.messages[:name].first}"
        errors.add(:rows, message)
      end
    end

    def valid_resource
      fetched_rows.each do |row_data|
        row_data.slice(:series).each do |_, data|
          next if data[:id].present?
          i18n_path = "activemodel.errors.forms.db/character_rows_form.invalid"
          errors.add(:rows, I18n.t(i18n_path, value: data[:value]))
        end
      end
    end
  end
end
