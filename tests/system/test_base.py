from dbeat import BaseTest

import os


class Test(BaseTest):

    def test_base(self):
        """
        Basic test with exiting dbeat normally
        """
        self.render_config_template(
            path=os.path.abspath(self.working_dir) + "/log/*"
        )

        dbeat_proc = self.start_beat()
        self.wait_until(lambda: self.log_contains("dbeat is running"))
        exit_code = dbeat_proc.kill_and_wait()
        assert exit_code == 0
