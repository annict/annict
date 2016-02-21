class PeopleController < ApplicationController
  def show(id)
    @person = Person.published.find(id)

    if @person.voice_actor?
      @casts_with_year = @person.
        casts.
        includes(work: [:season, :item]).
        group_by { |cast| cast.work.season.year }
      @cast_years = @casts_with_year.keys.sort.reverse
    end

    if @person.staff?
      @staffs_with_year = @person.
        staffs.
        includes(work: [:season, :item]).
        group_by { |staff| staff.work.season.year }
      @staff_years = @staffs_with_year.keys.sort.reverse
    end
  end
end
