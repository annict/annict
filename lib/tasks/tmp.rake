namespace :tmp do
  task update_seasons: :environment do
    years = 1917..1999
    seasons = [
      { name: "冬季", slug: "winter" },
      { name: "春季", slug: "spring" },
      { name: "夏季", slug: "summer" },
      { name: "秋季", slug: "autumn" }
    ]

    i = 1
    years.each do |y|
      seasons.each do |s|
        name = "#{y}年#{s[:name]}"
        Season.where(name: name).first_or_create(slug: "#{y}-#{s[:slug]}", sort_number: i)
        puts name
        i += 1
      end
    end

    Season.where(sort_number: nil).find_each do |s|
      s.update_column(:sort_number, i)
      i += 1
    end

    new_seasons = [
      { name: "2016年春季", slug: "2016-spring", sort_number: 398 },
      { name: "2016年夏季", slug: "2016-summer", sort_number: 399 },
      { name: "2016年秋季", slug: "2016-autumn", sort_number: 400 }
    ]
    new_seasons.each do |s|
      Season.create(name: s[:name], slug: s[:slug], sort_number: s[:sort_number])
    end
  end

  task update_work_season: :environment do
    Work.find_each do |w|
      next if w.season.present? || w.released_at.blank?

      year = w.released_at.year
      slug = case w.released_at.month
      when 1..3 then "#{year}-winter"
      when 4..6 then "#{year}-spring"
      when 7..9 then "#{year}-summer"
      when 10..12 then "#{year}-autumn"
      end
      season = Season.find_by(slug: slug)
      w.update_column(:season_id, season.id)
      puts "work: #{w.id}"
    end
  end
end
