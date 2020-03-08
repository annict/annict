# frozen_string_literal: true

module DBCommentDecorator
  def detail_url
    case model.class.name
    when "DBComment"
      case resource_type
      when "Work"
        "/works/#{resource.id}/activities##{anchor}"
      end
    end
  end
end
