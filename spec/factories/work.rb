FactoryGirl.define do
  factory :work do
    sequence(:title) { |n| "#{n}人はプリキュア" }
    media :tv
    official_site_url 'http://example.com'
    wikipedia_url 'http://example.com'

    trait :with_item do
      after :create do |work|
        create(:item, work: work)
      end
    end

    trait :with_current_season do
      after :create do |work|
        season = create(:season, slug: ENV["ANNICT_CURRENT_SEASON"])
        work.update(season_id: season.id)
      end
    end

    trait :with_next_season do
      after :create do |work|
        season = create(:season, slug: ENV["ANNICT_NEXT_SEASON"])
        work.update(season_id: season.id)
      end
    end

    trait :with_prev_season do
      after :create do |work|
        season = create(:season, slug: ENV["ANNICT_PREVIOUS_SEASON"])
        work.update(season_id: season.id)
      end
    end

    trait :with_episode do
      after :create do |work|
        create(:episode, work: work)
      end
    end
  end
end
