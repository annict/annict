# frozen_string_literal: true

DBPolicy = Struct.new(:user, :db) do
  def index?
    true
  end

  def show?
    true
  end

  def new?
    create?
  end

  def create?
    !!user&.committer?
  end

  def edit?
    update?
  end

  def update?
    !!user&.committer?
  end

  def appear?
    !!user&.committer?
  end

  def disappear?
    appear?
  end

  def destroy?
    !!user&.admin?
  end
end
