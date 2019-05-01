# frozen_string_literal: true

namespace :tmp do
  task fix_asin_on_work_image: :environment do
    WorkImage.where.not(asin: "").find_each do |wi|
      next if wi.asin.match?(/\A[0-9A-Z]{10}\z/)
      print "--- work_image: #{wi.id} "
      asin = wi.asin.scan(%r{(product|dp)/([0-9A-Z]{10})}).flatten[1]
      if asin&.match?(/\A[0-9A-Z]{10}\z/)
        puts "- #{asin}"
        wi.update_column(:asin, asin)
      else
        puts "- asin not found: #{wi.asin}"
      end
    end
  end

  task :paperclip_to_shrine, %i(from) => :environment do |_, args|
    from = Time.zone.parse(args[:from]).beginning_of_day

    WorkImage.where("updated_at >= ?", from).find_each do |wi|
      puts "--- work_image: #{wi.id}"
      imgix_url = "https://#{ENV.fetch('IMGIX_SOURCE')}"
      image_url = "#{imgix_url}/#{wi.attachment.path(:original)}"
      wi.file = Down.open(image_url)
      wi.save!
    end
  end
end
