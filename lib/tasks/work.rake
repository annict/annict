namespace :work do
  task notify_untouched_works: :environment do
    works = Work.where(episodes_count: 0).order(watchers_count: :desc).limit(3)
    WorkMailer.delay.untouched_works_notification(works.pluck(:id))
  end
end
