class SearchService
  attr_reader :q

  def initialize(q)
    @q = q
  end

  def works
    Work.search(title_cont: q).result
  end

  def people
    Person.search(name_cont: q).result
  end
end
