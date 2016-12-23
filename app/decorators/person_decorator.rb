# frozen_string_literal: true

class PersonDecorator < ApplicationDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || self.name
    h.link_to name, h.edit_db_person_path(self), options
  end

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
    h.link_to "@#{twitter_username}", url, target: "_blank"
  end

  def wikipedia_url_link
    h.link_to "Wikipedia", wikipedia_url, target: "_blank"
  end

  def url_link
    h.link_to URI.parse(url).host.downcase, url, target: "_blank"
  end

  def background_color
    miyamori = "#d8a57a"
    zuka = "#ae454c"
    miyazuka = "#c57965" # miyamori, zukaとの中間色

    if voice_actor? && staff?
      miyazuka
    elsif voice_actor?
      zuka
    elsif staff?
      miyamori
    end
  end

  def to_values
    model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :prefecture_id
        prefecture_id = send(:prefecture_id)
        Prefecture.find(prefecture_id).name if prefecture_id.present?
      when :gender
        gender = send(:gender)
        self.class.gender.find_value(gender).text if gender.present?
      when :url
        url = send(:url)
        h.link_to(url, url, target: "_blank") if url.present?
      when :media
        Work.media.find_value(send(:media)).text
      when :wikipedia_url
        wikipedia_url = send(field)
        if wikipedia_url.present?
          h.link_to(URI.decode(wikipedia_url), wikipedia_url, target: "_blank")
        end
      when :twitter_username
        username = send(:twitter_username)
        if username.present?
          url = "https://twitter.com/#{username}"
          h.link_to("@#{username}", url, target: "_blank")
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
