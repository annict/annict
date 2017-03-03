# frozen_string_literal: true

module DB
  class CharacterRowsForm
    include ActiveModel::Model
    include Virtus.model
    include ResourceRows

    attribute :rows, String

    validates :rows, presence: true
    validate :valid_character

    def save!
      new_characters_with_user.each(&:save_and_create_activity!)
    end

    private

    def attrs_list
      @attrs_list ||= parsed_rows.map do |row_columns|
        {
          name: row_columns[0]
        }
      end
    end

    def new_characters
      @new_characters ||= attrs_list.map { |attrs| Character.new(attrs) }
    end

    def new_characters_with_user
      @new_characters_with_user ||= new_characters.map do |character|
        character.user = @user
        character
      end
    end

    def valid_character
      return if new_characters.all?(&:valid?)

      new_characters.each do |c|
        next if c.valid?
        message = "\"#{c.name}\"#{c.errors.messages[:name].first}"
        errors.add(:rows, message)
      end
    end
  end
end
