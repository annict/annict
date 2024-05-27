# frozen_string_literal: true

module PersonOrgDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def local_name
      return name if I18n.locale == :ja
      return name_en if name_en.present?
      name
    end

    def local_other_name
      if I18n.locale == :ja
        name_kana.presence || ""
      else
        return name_kana if name_en.blank?
        name
      end
    end

    def name_with_other_name
      return "#{local_name} (#{local_other_name})" if local_other_name.present?
      local_name
    end

    def twitter_username_link
      url = "https://twitter.com/#{twitter_username}"
      link_to "@#{twitter_username}", url, target: "_blank", rel: "noopener"
    end
  end
end
