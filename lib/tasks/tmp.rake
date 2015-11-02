namespace :tmp do
  task update_seasons: :environment do
    Season.find_each do |s|
      year, name = s.slug.split("-")
      s.update(year: year, name: name)
      puts "season: #{s.slug}"
    end
  end
end
