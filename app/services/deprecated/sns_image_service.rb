# typed: false
# frozen_string_literal: true

class Deprecated::SnsImageService
  include HTTParty

  def initialize(work)
    @work = work
  end

  def save_og_image
    save_sns_image(:facebook_og_image_url, "og:image")
  end

  def save_twitter_image
    save_sns_image(:twitter_image_url, "twitter:image")
  end

  private

  def fetch_official_site
    self.class.get(@work.official_site_url)
  end

  def save_sns_image(column, property)
    return if @work.official_site_url.blank?

    begin
      html = fetch_official_site
      result = Nokogiri::HTML(html)
      image_html = result.css("meta[property='#{property}']")

      return if image_html.blank?

      @work.update_column(column, image_html.attr("content").to_s)
    rescue
      puts "error!: #{@work.id}"
    end
  end
end
