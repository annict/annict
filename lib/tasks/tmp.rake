# frozen_string_literal: true

namespace :tmp do
  task update_number_on_programs: :environment do
    works = Work.published
    works.find_each do |w|
      next if w.programs.empty?

      puts "--- work: #{w.id}"

      programs = w.programs.published.from("
        (
          select *,
            row_number() over (
              partition by channel_id
              order by started_at asc
            ) as row_num
          from programs
          where programs.work_id = #{w.id} and programs.aasm_state = 'published'
        ) as programs
      ")

      programs.each do |p|
        puts "--- work: #{w.id} program: #{p.id}"

        p.update_column(:number, p.row_num)
      end
    end
  end
end
