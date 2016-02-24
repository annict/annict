module PersonableDecorator
  extend ActiveSupport::Concern

  included do
    def name_with_old
      return name if name == person.name
      "#{name} (#{person.name})"
    end

    def name_with_old_link
      h.link_to name_with_old, h.person_path(person)
    end
  end
end
