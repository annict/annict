# frozen_string_literal: true

module CastDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || character&.name.presence || id
    link_to name, edit_db_cast_path(self), options
  end

  def local_name_with_old
    return local_name if local_name == person.local_name
    "#{local_name} (#{person.local_name})"
  end

  def local_name_with_old_link
    link_to local_name_with_old, person_path(person)
  end

  def to_values
    self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :person_id
        person_id = send(:person_id)
        Person.find(person_id).name if person_id.present?
      when :sort_number
        send(:sort_number).to_s
      when :character_id
        character_id = send(:character_id)
        Character.find(character_id).name if character_id.present?
      else
        send(field)
      end

      hash
    end
  end
end
