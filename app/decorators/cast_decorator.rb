# frozen_string_literal: true

class CastDecorator < ApplicationDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || name
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_work_cast_path(work, self)
    else
      h.new_db_work_draft_cast_path(work, cast_id: id)
    end

    h.link_to name, path, options
  end

  def name_with_old
    return name if name == person.name
    "#{name} (#{person.name})"
  end

  def name_with_old_link
    h.link_to name_with_old, h.person_path(person)
  end

  def to_values
    model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :person_id
        person_id = send(:person_id)
        Person.find(person_id).name if person_id.present?
      when :sort_number
        send(:sort_number).to_s
      when :character_id
        character_id = send(:character_id)
        Character.find(character_id).name if character_id.present?
      when :sort_number
      else
        send(field)
      end

      hash
    end
  end
end
