# frozen_string_literal: true

module CastStructDecorator
  def local_accurated_name
    return local_name if local_name == person.local_name
    "#{local_name} (#{person.local_name})"
  end

  def local_accurated_name_link
    link_to local_accurated_name, person_path(person.annict_id)
  end
end
