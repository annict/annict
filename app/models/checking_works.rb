class CheckingWorks
  include Enumerable

  def initialize(checking_works = [])
    @works = []

    checking_works.each do |work|
      work = CheckingWork.new(work) if work.instance_of?(Hash)
      add(work)
    end
  end

  def add(work)
    @works.push(work) unless work_ids.include?(work.id)
  end

  def work_ids
    @works.map { |work| work.id }
  end

  def models
    Work.where(id: @work_ids)
  end

  def to_a
    @works.map { |work| work.to_h }
  end

  def each
    @works.each do |work|
      yield work
    end
  end
end
