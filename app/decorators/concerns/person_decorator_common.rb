module PersonDecoratorCommon
  extend ActiveSupport::Concern

  included do
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
end
