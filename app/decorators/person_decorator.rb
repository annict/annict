# frozen_string_literal: true

module PersonDecorator
  include PersonOrgDecoratorCommon

  def name_link
    link_to local_name, person_path(self)
  end

  def db_detail_link(options = {})
    name = options.delete(:name).presence || self.name
    link_to name, db_edit_person_path(self), options
  end

  def grid_description(resource)
    case resource.class.name
    when "Cast"
      resource.character.decorate.name_link
    when "Staff"
      resource.decorate.role_name
    end
  end

  def to_values
    self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :prefecture_id
        prefecture_id = send(:prefecture_id)
        Prefecture.find(prefecture_id).name if prefecture_id.present?
      when :gender
        gender = send(:gender)
        self.class.gender.find_value(gender).text if gender.present?
      when :url
        url = send(:url)
        link_to(url, url, target: "_blank") if url.present?
      when :media
        Anime.media.find_value(send(:media)).text
      when :wikipedia_url
        wikipedia_url = send(field)
        if wikipedia_url.present?
          link_to(wikipedia_url, wikipedia_url, target: "_blank")
        end
      when :twitter_username
        username = send(:twitter_username)
        if username.present?
          url = "https://twitter.com/#{username}"
          link_to("@#{username}", url, target: "_blank")
        end
      when :birthday
        birthday = send(:birthday)
        birthday.strftime("%Y年%m月%d日") if birthday.present?
      when :blood_type
        blood_type = send(:blood_type)
        self.class.blood_type.find_value(blood_type).text if blood_type.present?
      when :height
        height = send(:height)
        "#{height} cm" if height.present?
      else
        send(field)
      end

      hash
    end
  end
end
