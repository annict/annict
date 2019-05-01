# frozen_string_literal: true

namespace :tmp do
  task fix_asin_on_work_image: :environment do
    WorkImage.where.not(asin: "").find_each do |wi|
      next if wi.asin.match?(/\A[0-9A-Z]{10}\z/)
      print "--- work_image: #{wi.id} "
      asin = wi.asin.scan(%r{(product|dp)/([0-9A-Z]{10})}).flatten[1]
      if asin && asin.match?(/\A[0-9A-Z]{10}\z/)
        puts "- #{asin}"
        wi.update_column(:asin, asin)
      else
        puts "- asin not found: #{wi.asin}"
      end
    end
  end
end
