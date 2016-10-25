# frozen_string_literal: true

class DbCommentDecorator < ApplicationDecorator
  def detail_url
    url = case model.class.name
    when "DbComment"
      ENV.fetch("ANNICT_DB_URL")
    else
      ENV.fetch("ANNICT_URL")
    end

    path = case model.class.name
    when "DbComment"
      case resource_type
      when "Work"
        "/works/#{resource.id}/activities##{anchor}"
      end
    end

    "#{url}/#{path}"
  end
end
