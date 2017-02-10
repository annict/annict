# frozen_string_literal: true

class DbCommentDecorator < ApplicationDecorator
  def detail_url
    case model.class.name
    when "DbComment"
      case resource_type
      when "Work"
        "/works/#{resource.id}/activities##{anchor}"
      end
    end
  end
end
