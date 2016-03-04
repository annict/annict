module PersonableDecorator
  extend ActiveSupport::Concern

  included do
    def name_with_old
      return name if name == resource.name
      "#{name} (#{resource.name})"
    end

    def name_with_old_link
      h.link_to name_with_old, h.person_path(resource)
    end
  end
end
