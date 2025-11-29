# typed: false
# frozen_string_literal: true

namespace :work_image do
  task save_sns_image: :environment do
    Work.only_kept.find_each do |work|
      puts "work: #{work.id}"
      service = Deprecated::SnsImageService.new(work)
      service.save_og_image
      service.save_twitter_image
    end
  end

  task save_recommended_image: :environment do
    Work.only_kept.find_each do |work|
      puts "work: #{work.id}"

      image_urls = {}
      max_dimension = 0
      recommended_image_url = ""

      if work.facebook_og_image_url.present?
        image_urls[:og_image] = work.facebook_og_image_url
      end

      if work.twitter_image_url.present?
        image_urls[:twitter_image] = work.twitter_image_url
      end

      if work.twitter_avatar_url.present?
        image_urls[:avatar] = work.twitter_avatar_url(:bigger)
      end

      image_urls.each do |key, image_url|
        image = MiniMagick::Image.open(image_url)
        dimension = image.dimensions.inject(1, :*)
        recommended_image_url = image_url if dimension > max_dimension
        max_dimension = dimension
      rescue
        case key
        when :og_image
          work.update_column(:facebook_og_image_url, "")
        when :twitter_image
          work.update_column(:twitter_image_url, "")
        end
      end

      work.update_column(:recommended_image_url, recommended_image_url)
    end
  end
end
