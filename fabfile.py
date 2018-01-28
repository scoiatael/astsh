from __future__ import with_statement
from fabric.api import run, env, settings
from fabric.context_managers import path, hide
from fabric.contrib.files import exists, upload_template

env.use_ssh_config = True

class Junest:
    def __init__(self):
        self.junest_dir = "$HOME/.local/share/junest"
        self.junest_bin = self.junest_dir + "/bin"

    def clone_repo(self):
        repo = "git://github.com/fsquillace/junest"
        dst = self.junest_dir
        if not exists(dst):
            clone_cmd = "git clone -q --single-branch --depth 1 {} {}".format(repo, dst)
            run(clone_cmd)

    def initialize_image(self):
        cmd = "pacman -Syy"
        self.run(cmd)

    def setup(self):
        self.clone_repo()
        self.initialize_image()

    def run(self, cmd):
        wrapped = "junest -f {}".format(cmd)
        with path(self.junest_bin):
            run(wrapped)

    def install(self, *pkgs):
        pkgs_str = " ".join(pkgs)
        cmd = "pacman --noprogressbar --needed -Syu {}".format(pkgs_str)
        self.run(cmd)

    def context(self):
        return {
            "path": self.junest_bin,
            "exec": "junest",
        }

def setup():
    adapter = Junest()
    adapter.setup()
    adapter.install("vim", "ranger", "fish")
    upload_template("activate", "~/activate",
                    context=adapter.context(),
                    use_jinja=True)
    run("chmod +x activate")
