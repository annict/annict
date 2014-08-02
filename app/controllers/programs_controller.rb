class ProgramsController < ApplicationController
  before_filter :authenticate_user!
end