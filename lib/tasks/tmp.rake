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
    errors = []

    [
      { model: WorkImage, paperclip_field: :attachment },
      { model: UserlandProject, paperclip_field: :icon },
      { model: Pv, paperclip_field: :thumbnail },
      { model: Item, paperclip_field: :thumbnail }
    ].each do |data|
      resources = data[:model].
        where("updated_at >= ?", from).
        where(image_data: nil)
      resources.find_each do |r|
        next if r.send(data[:paperclip_field]).blank?
        puts "--- #{data[:model].name}: #{r.id}"

        image_url = "#{ENV.fetch('ANNICT_FILE_STORAGE_URL')}/#{r.send(data[:paperclip_field]).path(:original)}"

        unless r.update(image: Down.open(image_url))
          error = { model: data[:model].name, id: r.id, messages: r.errors.messages }
          puts "--- Error!: #{error}"
          errors << error
        end
      end
    end

    puts errors
  end

  task :paperclip_to_shrine_profile, %i(from) => :environment do |_, args|
    from = Time.zone.parse(args[:from]).beginning_of_day
    errors = []

    [
      { paperclip_field: :tombo_avatar, shrine_field: :image },
      { paperclip_field: :tombo_background_image, shrine_field: :background_image }
    ].each do |data|
      resources = Profile.
        where("updated_at >= ?", from).
        where("#{data[:shrine_field]}_data": nil)
      resources.find_each do |r|
        next if r.send(data[:paperclip_field]).blank?
        puts "--- Profile: #{r.id}"

        image_url = "#{ENV.fetch('ANNICT_FILE_STORAGE_URL')}/#{r.send(data[:paperclip_field]).path(:original)}"
        unless r.update(data[:shrine_field] => Down.open(image_url))
          error = { model: "Profile", id: r.id, messages: r.errors.messages }
          puts "--- Error!: #{error}"
          errors << error
        end
      end
    end

    puts errors
  end
end
